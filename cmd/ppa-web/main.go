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

	// Original main() content goes here
	srv := server.NewServer()

	// Serve static files
	fs := http.FileServer(http.Dir("cmd/ppa-web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Main page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := templates.Index(srv.GetState()).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Get status
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		err := templates.StatusBar(srv.GetState()).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Set destination IP
	http.HandleFunc("/set-ip", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ip := r.FormValue("ip")
		if err := srv.ConnectToDevice(ip); err != nil {
			srv.SetState(func(state *server.AppState) {
				state.Status = "Error: " + err.Error()
				state.DestIP = ""
			})
		} else {
			srv.SetState(func(state *server.AppState) {
				state.DestIP = ip
				state.Status = "Connecting..."
			})
		}

		err := templates.IPForm(srv.GetState()).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Recall preset command
	http.HandleFunc("/recall", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if !srv.IsConnected() {
			http.Error(w, "Not connected to device", http.StatusBadRequest)
			return
		}

		presetNum := r.FormValue("preset")
		srv.LogPacket("Recalling preset %s", presetNum)
		// TODO: Implement preset recall using the client

		err := templates.LogWindow(srv.GetState()).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Set volume command
	http.HandleFunc("/volume", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if !srv.IsConnected() {
			http.Error(w, "Not connected to device", http.StatusBadRequest)
			return
		}

		volume := r.FormValue("volume")
		srv.LogPacket("Setting volume to %s", volume)
		// TODO: Implement volume control using the client

		err := templates.LogWindow(srv.GetState()).Render(r.Context(), w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

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
