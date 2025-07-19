package router

import (
	"net/http"
	"ppa-control/cmd/ppa-web/handler"

	"github.com/gorilla/mux"
)

// NewRouter creates a new router with all routes configured
func NewRouter(h *handler.Handler) *mux.Router {
	r := mux.NewRouter()

	// Static files - serve from static directory
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("cmd/ppa-web/static"))))

	// Built static files - serve from static-dist directory
	r.PathPrefix("/static-dist/").Handler(http.StripPrefix("/static-dist/", http.FileServer(http.Dir("cmd/ppa-web/static-dist"))))

	// Main routes
	r.HandleFunc("/", h.HandleIndex).Methods(http.MethodGet)
	r.HandleFunc("/docs", h.HandleDocs).Methods(http.MethodGet)

	// SPA route for development (serve dev.html which will load the modules)
	r.HandleFunc("/spa", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "cmd/ppa-web/static/dev.html")
	}).Methods(http.MethodGet)

	// Production SPA route (serve built React app)
	r.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "cmd/ppa-web/static-dist/index.html")
	}).Methods(http.MethodGet)
	r.HandleFunc("/status", h.HandleStatus).Methods(http.MethodGet)
	r.HandleFunc("/set-ip", h.HandleSetIP).Methods(http.MethodPost)
	r.HandleFunc("/recall", h.HandleRecall).Methods(http.MethodPost)
	r.HandleFunc("/volume", h.HandleVolume).Methods(http.MethodPost)

	// Discovery routes
	r.HandleFunc("/discovery/start", h.HandleStartDiscovery).Methods(http.MethodPost)
	r.HandleFunc("/discovery/stop", h.HandleStopDiscovery).Methods(http.MethodPost)
	r.HandleFunc("/discovery/events", h.HandleDiscoveryEvents).Methods(http.MethodGet)

	// API routes for packet analysis
	api := r.PathPrefix("/api").Subrouter()

	// PCAP file management
	api.HandleFunc("/pcap/upload", h.HandleUploadPCAP).Methods(http.MethodPost)
	api.HandleFunc("/pcap/{id}/analyze", h.HandleAnalyzePCAP).Methods(http.MethodPost)
	api.HandleFunc("/pcap/{id}/status", h.HandlePCAPStatus).Methods(http.MethodGet)
	api.HandleFunc("/pcap/list", h.HandleListPCAPs).Methods(http.MethodGet)

	// Analysis results
	api.HandleFunc("/analysis/{id}", h.HandleGetAnalysisResult).Methods(http.MethodGet)
	api.HandleFunc("/analysis/list", h.HandleListAnalysisResults).Methods(http.MethodGet)

	// Document API routes
	api.HandleFunc("/docs/search", h.HandleDocumentSearch).Methods(http.MethodGet)
	api.HandleFunc("/docs/list", h.HandleDocumentList).Methods(http.MethodGet)
	api.HandleFunc("/docs/view", h.HandleDocumentView).Methods(http.MethodGet)
	api.HandleFunc("/markdown/render", h.HandleMarkdownRender).Methods(http.MethodPost)

	// Legacy routes for backward compatibility
	api.HandleFunc("/captures", h.HandleListCaptures).Methods(http.MethodGet)
	api.HandleFunc("/analyze/{filename}", h.HandleAnalyze).Methods(http.MethodPost)
	api.HandleFunc("/analysis/{session}", h.HandleGetAnalysis).Methods(http.MethodGet)
	api.HandleFunc("/files/{path:.*}", h.HandleServeFile).Methods(http.MethodGet)
	api.HandleFunc("/search", h.HandleSearch).Methods(http.MethodGet)
	api.HandleFunc("/packets/{session}", h.HandleGetPackets).Methods(http.MethodGet)

	// Add CORS middleware for API routes
	api.Use(corsMiddleware)

	return r
}

// corsMiddleware adds CORS headers for development
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
