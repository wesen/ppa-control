package handler

import (
	"net/http"
	"ppa-control/cmd/ppa-web/server"
	"ppa-control/cmd/ppa-web/templates"
)

// Handler encapsulates all HTTP handlers for the web interface
type Handler struct {
	srv *server.Server
}

// NewHandler creates a new Handler instance
func NewHandler(srv *server.Server) *Handler {
	return &Handler{srv: srv}
}

// HandleIndex handles the main page request
func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	err := templates.Index(h.srv.GetState()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleStatus handles status update requests
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	err := templates.StatusBar(h.srv.GetState()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleSetIP handles device connection requests
func (h *Handler) HandleSetIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ip := r.FormValue("ip")
	if err := h.srv.ConnectToDevice(ip); err != nil {
		h.srv.SetState(func(state *server.AppState) {
			state.Status = "Error: " + err.Error()
			state.DestIP = ""
		})
	} else {
		h.srv.SetState(func(state *server.AppState) {
			state.DestIP = ip
			state.Status = "Connecting..."
		})
	}

	err := templates.IPForm(h.srv.GetState()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleRecall handles preset recall requests
func (h *Handler) HandleRecall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.srv.IsConnected() {
		http.Error(w, "Not connected to device", http.StatusBadRequest)
		return
	}

	presetNum := r.FormValue("preset")
	h.srv.LogPacket("Recalling preset %s", presetNum)
	// TODO: Implement preset recall using the client

	err := templates.LogWindow(h.srv.GetState()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleVolume handles volume control requests
func (h *Handler) HandleVolume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.srv.IsConnected() {
		http.Error(w, "Not connected to device", http.StatusBadRequest)
		return
	}

	volume := r.FormValue("volume")
	h.srv.LogPacket("Setting volume to %s", volume)
	// TODO: Implement volume control using the client

	err := templates.LogWindow(h.srv.GetState()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleStartDiscovery handles starting the discovery process
func (h *Handler) HandleStartDiscovery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.srv.StartDiscovery(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := templates.DiscoverySection(h.srv.GetState()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleStopDiscovery handles stopping the discovery process
func (h *Handler) HandleStopDiscovery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.srv.StopDiscovery(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err := templates.DiscoverySection(h.srv.GetState()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleDiscoveryEvents handles SSE for discovery updates
func (h *Handler) HandleDiscoveryEvents(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create notification channel
	updateCh := make(chan struct{})
	h.srv.AddUpdateListener(updateCh)
	defer h.srv.RemoveUpdateListener(updateCh)

	// Send initial state
	err := templates.DiscoveredDevices(h.srv.GetState().DiscoveredDevices).Render(r.Context(), w)
	if err != nil {
		return
	}
	w.(http.Flusher).Flush()

	// Send updates
	for {
		select {
		case <-r.Context().Done():
			return
		case <-updateCh:
			err := templates.DiscoveredDevices(h.srv.GetState().DiscoveredDevices).Render(r.Context(), w)
			if err != nil {
				return
			}
			w.(http.Flusher).Flush()
		}
	}
}
