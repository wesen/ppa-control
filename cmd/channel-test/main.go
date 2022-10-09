package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"math/rand"
	"time"
)

// I never remember how go drains channels, fucking annoying

// I always get confused in [[golang]] on how to close channels when a writer might have the receiver exit midway.
// Buffered channels is not really the way to go here since
// I don't know how big the amount of added entries is going to be. Instead, by sending inside a select,
// I make sure that the context is not done before writing, in which case I might block.
// But still, does that work? I might still be one late, right?
//
// I think this has a subtle race condition, actually.
// What if context is cancelled, the writer has already written to the channel and is hanging,
// and the receiver exits... I need to make the channel buffered, I think. I'm gonna read up in
// [[BOOK - 100 Go Mistakes and How to Avoid Them - Teiva Harsanyi]] to see if that problem is addressed.
// Otherwise it surely is addressed in [[ZK - golang Concurrency is not easy]].

func createStrings(ctx context.Context, stringCh chan string) error {
	counter := 0
	for {
		t := time.NewTimer(1 * time.Second)

		sendStrings := func() {
			randInt := rand.Intn(5) + 5
			for i := 0; i < randInt; i++ {
				select {
				case <-ctx.Done():
					return
				case stringCh <- fmt.Sprintf("cnt %d", counter):
				}
				counter++
			}
		}

		sendStrings()

		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			sendStrings()
		}
	}
}

func poller(ctx context.Context, stringCh chan string) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case s := <-stringCh:
			fmt.Println(s)
			if s == "cnt 23" {
				return fmt.Errorf("got 23")
			}
		}
	}
}

func main() {
	grp, ctx := errgroup.WithContext(context.Background())
	stringCh := make(chan string)

	grp.Go(func() error {
		return createStrings(ctx, stringCh)
	})
	grp.Go(func() error {
		return poller(ctx, stringCh)
	})

	err := grp.Wait()

	// drain stringCh
drainStringCh:
	for {
		select {
		case <-stringCh:
		default:
			break drainStringCh
		}
	}

	if err != nil {
		fmt.Println(err)
	}
}
