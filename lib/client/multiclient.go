package client

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
	"go.uber.org/atomic"
)

type MultiClient struct {
	wg   sync.WaitGroup
	name string

	// maps from address to pkg
	mutex   sync.RWMutex
	clients map[string]Client
	cancels map[string]context.CancelFunc

	receivedCh chan ReceivedMessage
	errorCh    chan error // Channel for error propagation

	waiting atomic.Bool
}

func NewMultiClient(name string) *MultiClient {
	return &MultiClient{
		name:       name,
		clients:    make(map[string]Client),
		cancels:    make(map[string]context.CancelFunc),
		receivedCh: make(chan ReceivedMessage, 10),
		errorCh:    make(chan error, 10), // Buffer for errors
		waiting:    *atomic.NewBool(false),
	}
}

// Commander interface implementation
func (mc *MultiClient) SendPing() {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	for addr, c := range mc.clients {
		if err := mc.safeSend(addr, c.SendPing); err != nil {
			log.Error().Err(err).Str("addr", addr).Msg("failed to send ping")
		}
	}
}

func (mc *MultiClient) SendPresetRecallByPresetIndex(index int) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	for addr, c := range mc.clients {
		if err := mc.safeSend(addr, func() { c.SendPresetRecallByPresetIndex(index) }); err != nil {
			log.Error().Err(err).Str("addr", addr).Msg("failed to send preset recall")
		}
	}
}

func (mc *MultiClient) SendMasterVolume(volume float32) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	for addr, c := range mc.clients {
		if err := mc.safeSend(addr, func() { c.SendMasterVolume(volume) }); err != nil {
			log.Error().Err(err).Str("addr", addr).Msg("failed to send master volume")
		}
	}
}

// safeSend executes a send operation safely and returns any error
func (mc *MultiClient) safeSend(addr string, fn func()) error {
	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in send operation: %v", r)
			mc.errorCh <- NewClientError("send", addr, err)
		}
	}()
	fn()
	return nil
}

func (mc *MultiClient) DoesClientExist(addr string) bool {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	_, exists := mc.clients[addr]
	return exists
}

func (mc *MultiClient) AddClient(ctx context.Context, addrPort string, iface string, componentId uint) (Client, error) {
	if mc.waiting.Load() {
		return nil, &ErrClientBusy{Operation: "shutdown"}
	}

	log.Debug().
		Str("name", mc.name).
		Str("iface", iface).
		Str("addrPort", addrPort).
		Msg("adding client")

	if mc.DoesClientExist(addrPort) {
		return nil, &ErrClientExists{Addr: addrPort}
	}

	c := NewSingleDevice(addrPort, iface, componentId)
	clientCtx, cancel := context.WithCancel(ctx)
	func() {
		mc.mutex.Lock()
		defer mc.mutex.Unlock()
		mc.clients[addrPort] = c
		mc.cancels[addrPort] = cancel
	}()

	mc.wg.Add(1)
	go func() {
		defer mc.wg.Done()

		log.Info().
			Str("name", mc.name).
			Str("iface", iface).
			Str("addrPort", addrPort).
			Msg("starting client")

		if err := c.Run(clientCtx, mc.receivedCh); err != nil {
			mc.errorCh <- NewClientError("run", addrPort, err)
			log.Error().
				Str("name", mc.name).
				Str("iface", iface).
				Str("addrPort", addrPort).
				Err(err).
				Msg("client stopped with error")
		} else {
			log.Info().
				Str("name", mc.name).
				Str("iface", iface).
				Str("addrPort", addrPort).
				Msg("client stopped")
		}

		func() {
			mc.mutex.Lock()
			defer mc.mutex.Unlock()
			delete(mc.clients, addrPort)
			delete(mc.cancels, addrPort)
		}()
	}()

	return c, nil
}

func (mc *MultiClient) CancelClient(addr string) error {
	if mc.waiting.Load() {
		return &ErrClientBusy{Operation: "shutdown"}
	}

	log.Debug().
		Str("name", mc.name).
		Str("addr", addr).
		Msg("cancelling client")

	if !mc.DoesClientExist(addr) {
		return &ErrClientNotFound{Addr: addr}
	}

	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if cancel, exists := mc.cancels[addr]; exists {
		cancel()
		return nil
	}
	return &ErrClientNotFound{Addr: addr}
}

func (mc *MultiClient) Run(ctx context.Context, receivedCh chan<- ReceivedMessage) (err error) {
	// Error handling goroutine
	go func() {
		for err := range mc.errorCh {
			log.Error().Err(err).Msg("client error received")
			// Could add error handling strategy here
		}
	}()

	for {
		select {
		case m := <-mc.receivedCh:
			select {
			case receivedCh <- m:
			case <-ctx.Done():
				return ctx.Err()
			}

		case <-ctx.Done():
			log.Debug().
				Str("name", mc.name).
				Msg("context done, stopping multiclient")
			mc.waiting.Store(true)

			func() {
				mc.mutex.RLock()
				defer mc.mutex.RUnlock()

				for _, cancel := range mc.cancels {
					cancel()
				}
			}()

			mc.wg.Wait()
			close(mc.errorCh) // Close error channel after all clients are done

			log.Debug().
				Str("name", mc.name).
				Msg("multiclient stopped")

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
