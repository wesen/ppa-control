package main

import (
	"log"
	"net/http"
	"os"
	"ppa-control/cmd/ppa-web/server"
	"ppa-control/cmd/ppa-web/templates"

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
}

func run(cmd *cobra.Command, args []string) error {
	// Set up logging
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(level)

	// Create server and handler
	srv := server.NewServer()
	handler := server.NewHandler(srv, templates.NewTemplateProvider())

	// Serve static files
	fs := http.FileServer(http.Dir("cmd/ppa-web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Register handlers
	http.HandleFunc("/", handler.HandleIndex)
	http.HandleFunc("/status", handler.HandleStatus)
	http.HandleFunc("/set-ip", handler.HandleSetIP)
	http.HandleFunc("/recall", handler.HandleRecall)
	http.HandleFunc("/volume", handler.HandleVolume)

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
