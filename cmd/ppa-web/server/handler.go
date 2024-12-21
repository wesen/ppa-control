package server

import (
	"context"
	"net/http"
)

// TemplateRenderer is an interface for rendering templates
type TemplateRenderer interface {
	Render(ctx context.Context, w http.ResponseWriter) error
}

// TemplateProvider provides template rendering functions
type TemplateProvider interface {
	Index(state AppState) TemplateRenderer
	StatusBar(state AppState) TemplateRenderer
	IPForm(state AppState) TemplateRenderer
	LogWindow(state AppState) TemplateRenderer
}

// Handler encapsulates all HTTP handlers for the web interface
type Handler struct {
	srv       *Server
	templates TemplateProvider
}

// NewHandler creates a new Handler instance
func NewHandler(srv *Server, templates TemplateProvider) *Handler {
	return &Handler{
		srv:       srv,
		templates: templates,
	}
}

// HandleIndex handles the main page request
func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	err := h.templates.Index(h.srv.GetState()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleStatus handles status update requests
func (h *Handler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	err := h.templates.StatusBar(h.srv.GetState()).Render(r.Context(), w)
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
		h.srv.SetState(func(state *AppState) {
			state.Status = "Error: " + err.Error()
			state.DestIP = ""
		})
	} else {
		h.srv.SetState(func(state *AppState) {
			state.DestIP = ip
			state.Status = "Connecting..."
		})
	}

	err := h.templates.IPForm(h.srv.GetState()).Render(r.Context(), w)
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

	err := h.templates.LogWindow(h.srv.GetState()).Render(r.Context(), w)
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

	err := h.templates.LogWindow(h.srv.GetState()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
