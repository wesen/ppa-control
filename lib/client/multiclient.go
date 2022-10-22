package client

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.uber.org/atomic"
	"strings"
	"sync"
)

type MultiClient struct {
	wg sync.WaitGroup

	// maps from address to client
	mutex   sync.RWMutex
	clients map[string]Client
	cancels map[string]context.CancelFunc

	receivedCh chan ReceivedMessage

	waiting atomic.Bool
}

func NewMultiClient() *MultiClient {
	return &MultiClient{
		clients:    make(map[string]Client),
		cancels:    make(map[string]context.CancelFunc),
		receivedCh: make(chan ReceivedMessage, 10),
		waiting:    *atomic.NewBool(false),
	}
}

func (mc *MultiClient) SendPing() {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	for _, c := range mc.clients {
		c.SendPing()
	}
}

func (mc *MultiClient) SendPresetRecallByPresetIndex(index int) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	for _, c := range mc.clients {
		c.SendPresetRecallByPresetIndex(index)
	}
}

func (mc *MultiClient) DoesClientExist(addr string) bool {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	_, exists := mc.clients[addr]
	return exists
}

func (mc *MultiClient) AddClient(ctx context.Context, addr string, componentId uint) (Client, error) {
	if mc.waiting.Load() {
		panic("cannot add client while waiting for clients to be done")
	}

	log.Debug().Str("addr", addr).Msg("adding client")
	if mc.DoesClientExist(addr) {
		log.Error().Str("addr", addr).Msg("client already exists")
		return nil, fmt.Errorf("client for %s already exists", addr)
	}

	c := NewSingleDevice(addr, componentId)
	clientCtx, cancel := context.WithCancel(ctx)
	func() {
		mc.mutex.Lock()
		defer mc.mutex.Unlock()
		mc.clients[addr] = c
		mc.cancels[addr] = cancel
	}()

	go func() {
		mc.wg.Add(1)

		log.Info().Str("addr", addr).Msg("starting client")
		err := c.Run(clientCtx, &mc.receivedCh)

		func() {
			mc.mutex.Lock()
			defer mc.mutex.Unlock()
			delete(mc.clients, addr)
			delete(mc.cancels, addr)
		}()

		mc.wg.Done()
		if err != nil {
			log.Error().Str("addr", addr).Err(err).Msg("client stopped with error")
		} else {
			log.Info().Str("addr", addr).Msg("client stopped")
		}
	}()

	return c, nil
}

func (mc *MultiClient) CancelClient(addr string) error {
	if mc.waiting.Load() {
		panic("cannot remove client while waiting for clients to be done")
	}

	log.Debug().Str("addr", addr).Msg("adding client")
	if !mc.DoesClientExist(addr) {
		log.Error().Str("addr", addr).Msg("client does not exist")
		return fmt.Errorf("client for %s does not exist", addr)
	}

	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	mc.cancels[addr]()

	return nil
}

func (mc *MultiClient) Run(ctx context.Context, receivedCh *chan ReceivedMessage) (err error) {
	for {
		select {
		case m := <-mc.receivedCh:
			*receivedCh <- m

		case <-ctx.Done():
			log.Debug().Msg("context done, stopping multiclient")
			mc.waiting.Store(true)

			func() {
				mc.mutex.RLock()
				defer mc.mutex.RUnlock()

				for _, cancel := range mc.cancels {
					cancel()
				}
			}()

			mc.wg.Wait()
			log.Debug().Msg("multiclient stopped")

			// do I need to drain channels here?
			// All clients have been closed at this point, so no one is writing
			// to receivedCh, and if I close the channel with things buffered, and I'm the only
			// receiver, then things are fine too, I think.

			return ctx.Err()
		}
	}
}

func (mc *MultiClient) Name() string {
	var names []string
	for _, c := range mc.clients {
		names = append(names, c.Name())
	}
	return "multiclient-" + strings.Join(names, ",")
}
