package main

import (
	"context"
	"golang.org/x/sync/errgroup"
	"log"
	"ppa-control/lib/utils"
	"time"
)

// the scenario i think is problematic
// a function creates a struct instance
//   the struct instance contains a channel
//   we start a goroutine that writes multiple values to this channel on a select tick

type Client struct {
	ch chan int
}

func (c *Client) Run(ctx context.Context) error {
	t := time.NewTicker(1 * time.Millisecond)

	for {

		select {
		case <-ctx.Done():
			log.Println("C: pkg done")
			return ctx.Err()
		case <-t.C:
			for i := 0; i < 10; i++ {
				log.Println("C: waiting for pkg write")
				time.Sleep(1 * time.Second)
				log.Println("C: pkg write")
				select {
				case <-ctx.Done():
					log.Println("C: pkg done")
					return ctx.Err()
				case c.ch <- i:
					log.Println("C: wrote to pkg channel", i)
				}
			}
		}

		t.Reset(5 * time.Second)
	}
}

func run(ctx context.Context) error {
	client := &Client{ch: make(chan int)}

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		log.Println("R: starting pkg")
		err := client.Run(ctx)
		log.Println("R: pkg done")
		return err
	})

	grp.Go(func() error {
		for {
			log.Println("R: waiting for pkg read")
			select {
			case <-ctx.Done():
				log.Println("R: done")
				return ctx.Err()
			case v := <-client.ch:
				log.Println("R: read from pkg channel", v)
				println(v)
			}
		}
	})

	log.Println("R: waiting for group")
	return grp.Wait()
}

func main() {
	utils.StartBackgroundLeakTracker(2 * time.Second)
	utils.StartSIGPOLLStacktraceDumper("")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		log.Println("M: waiting for run")
		_ = run(ctx)
		log.Println("M: run done")
	}()

	time.Sleep(7 * time.Second)
	log.Println("M: cancel")
	cancel()
	log.Println("M: cancelled")

	time.Sleep(60 * time.Second)
}
