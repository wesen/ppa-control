package main

import (
	"log"
	"net/http"
	"os"
	"ppa-control/cmd/ppa-web/handler"
	"ppa-control/cmd/ppa-web/server"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	logLevel string
	rootCmd  = &cobra.Command{
		Use:   "ppa-web",
		Short: "Web interface for PPA control",
		RunE:  run,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "l", "debug", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolP("discover", "d", false, "Enable device discovery")
	rootCmd.PersistentFlags().StringArray("interfaces", []string{}, "Interfaces to use for discovery")
	rootCmd.PersistentFlags().UintP("port", "p", 5001, "Port to use for device communication")
}

func run(cmd *cobra.Command, args []string) error {
	// Set up logging
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(level)

	// Create server and handler
	srv := server.FromCobraCommand(cmd)
	handler := handler.NewHandler(srv)

	// Serve static files
	fs := http.FileServer(http.Dir("cmd/ppa-web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register handlers
	http.HandleFunc("/", handler.HandleIndex)
	http.HandleFunc("/status", handler.HandleStatus)
	http.HandleFunc("/set-ip", handler.HandleSetIP)
	http.HandleFunc("/recall", handler.HandleRecall)
	http.HandleFunc("/volume", handler.HandleVolume)
	http.HandleFunc("/discovery/start", handler.HandleStartDiscovery)
	http.HandleFunc("/discovery/stop", handler.HandleStopDiscovery)
	http.HandleFunc("/discovery/events", handler.HandleDiscoveryEvents)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
