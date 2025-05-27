package router

import (
	"net/http"
	"ppa-control/cmd/ppa-web/handler"

	"github.com/gorilla/mux"
)

// NewRouter creates a new router with all routes configured
func NewRouter(h *handler.Handler) *mux.Router {
	r := mux.NewRouter()

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("cmd/ppa-web/static"))))

	// Main routes
	r.HandleFunc("/", h.HandleIndex).Methods(http.MethodGet)
	r.HandleFunc("/status", h.HandleStatus).Methods(http.MethodGet)
	r.HandleFunc("/set-ip", h.HandleSetIP).Methods(http.MethodPost)
	r.HandleFunc("/recall", h.HandleRecall).Methods(http.MethodPost)
	r.HandleFunc("/volume", h.HandleVolume).Methods(http.MethodPost)

	// Discovery routes
	r.HandleFunc("/discovery/start", h.HandleStartDiscovery).Methods(http.MethodPost)
	r.HandleFunc("/discovery/stop", h.HandleStopDiscovery).Methods(http.MethodPost)
	r.HandleFunc("/discovery/events", h.HandleDiscoveryEvents).Methods(http.MethodGet)

	return r
}
