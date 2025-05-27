package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"ppa-control/cmd/ppa-web/handler"
	"ppa-control/cmd/ppa-web/router"
	"ppa-control/cmd/ppa-web/server"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

type statusResponseWriter struct {
	http.Flusher
	http.ResponseWriter
	status int
	size   int
}

func (w *statusResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a request ID and add it to the context
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		// Create a custom response writer to capture status code
		sw := &statusResponseWriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		// Recover from panics
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Str("request_id", requestID).
					Str("stack", string(debug.Stack())).
					Interface("error", err).
					Msg("Handler panic recovered")

				sw.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(sw, "Internal Server Error")
			}

			// Log the request details after completion
			log.Info().
				Str("request_id", requestID).
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Int("status", sw.status).
				Int("size", sw.size).
				Dur("duration", time.Since(start)).
				Msg("Request completed")
		}()

		next.ServeHTTP(sw, r)
	})
}

func run(cmd *cobra.Command, args []string) error {
	// Set up logging
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return err
	}

	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	log.Logger = zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	zerolog.SetGlobalLevel(level)

	// Create server and handler
	srv := server.FromCobraCommand(cmd)
	h := handler.NewHandler(srv)
	r := router.NewRouter(h)

	// Add middleware
	r.Use(loggingMiddleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Info().Msgf("Starting server on :%s", port)

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	return httpServer.ListenAndServe()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute root command")
		os.Exit(1)
	}
}
