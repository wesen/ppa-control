package client

import (
	"context"
	"golang.org/x/sync/errgroup"
	"strings"
)

type MultiClient struct {
	clients []Client
}

func NewMultiClient(clients []Client) *MultiClient {
	return &MultiClient{clients: clients}
}

func (mc *MultiClient) SendPing() {
	for _, c := range mc.clients {
		c.SendPing()
	}
}

func (mc *MultiClient) SendPresetRecallByPresetIndex(index int) {
	for _, c := range mc.clients {
		c.SendPresetRecallByPresetIndex(index)
	}
}

func (mc *MultiClient) Run(ctx context.Context, c *chan ReceivedMessage) (err error) {
	grp, ctx := errgroup.WithContext(ctx)

	for _, c2 := range mc.clients {
		c3 := c2
		grp.Go(func() error {
			return c3.Run(ctx, c)
		})
	}
	return grp.Wait()
}

func (mc *MultiClient) Name() string {
	var names []string
	for _, c := range mc.clients {
		names = append(names, c.Name())
	}
	return "multiclient-" + strings.Join(names, ",")
}