package utils

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"
)

func StartSIGPOLLStacktraceDumper(memProfileFile string) {
	// SIGPOLL will print out a stacktrace of running goroutines, and write a mem profile if memprofiling is enabled
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGHUP)
		buf := make([]byte, 1<<20)
		i := 0
		for {
			i += 1
			<-sigs
			stacklen := runtime.Stack(buf, true)
			// this is pretty rough, could probably make it more structured
			log.Printf("=== received SIGPOLL ===\n*** goroutine dump...\n%s\n*** end\n", buf[:stacklen])
			if memProfileFile != "" {
				memProfileFile := fmt.Sprintf("%s.%0.3d", memProfileFile, i)
				WriteMemprofile(memProfileFile)
				log.Printf("Logged mem profile to %s\n", memProfileFile)
			}
		}
	}()
}

func WriteMemprofile(memprofile string) {
	f, err := os.Create(memprofile)
	if err != nil {
		log.Error().Err(err).Msg("could not create memory profile")
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Error().Err(err).Msg("could not write memory profile")
	}
	_ = f.Close()
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func StartBackgroundLeakTracker(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			runtime.GC()
			fmt.Println()
			printMemUsage()
			fmt.Printf("======= Goroutines: %d\n\n", runtime.NumGoroutine())
		}
	}()
}
