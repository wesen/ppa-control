package handler

import (
	"bytes"
	"fmt"
	"net/http"
	"ppa-control/cmd/ppa-web/templates"
	"ppa-control/cmd/ppa-web/types"
	"time"

	"github.com/rs/zerolog/log"
)

// Handler encapsulates all HTTP handlers for the web interface
type Handler struct {
	srv types.ServerInterface
}

// NewHandler creates a new Handler instance
func NewHandler(srv types.ServerInterface) *Handler {
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
		h.srv.SetState(func(state *types.AppState) {
			state.Status = "Error: " + err.Error()
			state.DestIP = ""
		})
	} else {
		h.srv.SetState(func(state *types.AppState) {
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

// sseWriter wraps a ResponseWriter to provide SSE functionality
type sseWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// newSSEWriter creates a new SSE writer if possible
func newSSEWriter(w http.ResponseWriter) (*sseWriter, error) {
	// Try to get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming not supported")
	}

	return &sseWriter{
		w:       w,
		flusher: flusher,
	}, nil
}

// writeEvent writes an SSE event with the given event type and data
func (sw *sseWriter) writeEvent(eventType, data string) error {
	if _, err := fmt.Fprintf(sw.w, "event: %s\ndata: %s\n\n", eventType, data); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}
	sw.flusher.Flush()
	return nil
}

// writePing writes a ping event
func (sw *sseWriter) writePing() error {
	return sw.writeEvent("ping", "")
}

// writeMessage writes a message event with the given data
func (sw *sseWriter) writeMessage(data string) error {
	return sw.writeEvent("message", data)
}

// HandleDiscoveryEvents handles SSE for discovery updates
func (h *Handler) HandleDiscoveryEvents(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create SSE writer
	sseWriter, err := newSSEWriter(w)
	if err != nil {
		http.Error(w, "Streaming unsupported! The server does not support streaming responses.", http.StatusPreconditionFailed)
		return
	}

	// Create notification channel with buffer to prevent blocking
	updateCh := make(chan struct{}, 1)
	h.srv.AddUpdateListener(updateCh)
	defer h.srv.RemoveUpdateListener(updateCh)

	// Create done channel for cleanup
	done := make(chan bool)
	defer close(done)

	// Start heartbeat goroutine
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if err := sseWriter.writePing(); err != nil {
					log.Error().Err(err).Msg("Failed to write ping")
					return
				}
			}
		}
	}()

	// Send initial state
	var buf bytes.Buffer
	if err := templates.DiscoveredDevices(h.srv.GetState().DiscoveredDevices).Render(r.Context(), &buf); err != nil {
		http.Error(w, fmt.Sprintf("Failed to render template: %v", err), http.StatusInternalServerError)
		return
	}

	if err := sseWriter.writeMessage(buf.String()); err != nil {
		log.Error().Err(err).Msg("Failed to write initial state")
		return
	}

	// Send updates
	for {
		select {
		case <-r.Context().Done():
			return
		case <-updateCh:
			buf.Reset()
			if err := templates.DiscoveredDevices(h.srv.GetState().DiscoveredDevices).Render(r.Context(), &buf); err != nil {
				log.Error().Err(err).Msg("Failed to render update")
				return
			}
			if err := sseWriter.writeMessage(buf.String()); err != nil {
				log.Error().Err(err).Msg("Failed to write update")
				return
			}
		}
	}
}
