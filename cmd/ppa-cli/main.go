package main

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/sync/errgroup"
	"ppa-control/lib/client"
	"ppa-control/lib/server"
	"time"
)

var (
	address        = flag.String("address", "127.0.0.1", "server address")
	port           = flag.Uint("port", 5151, "server port")
	runServer      = flag.Bool("run-server", false, "Run as server too")
	presetPosition = flag.Int("position", 1, "preset")
	componentId    = flag.Int("component-id", 0xff, "component ID (default: 0xff)")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	serverString := fmt.Sprintf("%s:%d", *address, *port)
	fmt.Printf("Connecting to %s\n", serverString)

	if *runServer {
		fmt.Printf("Starting test server")
		go server.RunServer(ctx, serverString)
		time.Sleep(1 * time.Second)
	}

	c := client.NewClient(*componentId)

	go func() {
		for {
			c.SendPresetRecallByPresetIndex(*presetPosition)
			time.Sleep(1 * time.Second)
		}
	}()

	grp, ctx2 := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return c.Run(ctx2, serverString)
	})
	err := grp.Wait()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
