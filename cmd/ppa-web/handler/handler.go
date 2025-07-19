package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"ppa-control/cmd/ppa-web/templates"
	"ppa-control/cmd/ppa-web/types"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// Handler encapsulates all HTTP handlers for the web interface
type Handler struct {
	srv        types.ServerInterface
	docService *DocumentService
}

// NewHandler creates a new Handler instance
func NewHandler(srv types.ServerInterface) *Handler {
	return &Handler{
		srv:        srv,
		docService: nil, // Will be set later with SetDocumentService
	}
}

// SetDocumentService sets the document service for the handler
func (h *Handler) SetDocumentService(ds *DocumentService) {
	h.docService = ds
}

// HandleIndex handles the main page request
func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	err := templates.Index(h.srv.GetState()).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// HandleDocs handles the documents page request
func (h *Handler) HandleDocs(w http.ResponseWriter, r *http.Request) {
	err := templates.DocsPage().Render(r.Context(), w)
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

// API handlers for packet analysis

// PCAP file upload and management

// HandleUploadPCAP handles file uploads
func (h *Handler) HandleUploadPCAP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (32MB max memory)
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	if !strings.HasSuffix(header.Filename, ".pcap") &&
		!strings.HasSuffix(header.Filename, ".pcapng") &&
		!strings.HasSuffix(header.Filename, ".cap") {
		http.Error(w, "Invalid file type. Only .pcap, .pcapng, and .cap files are allowed", http.StatusBadRequest)
		return
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "ttmp/2025-07-13/pcap/uploads"
	os.MkdirAll(uploadsDir, 0755)

	// Generate unique filename
	id := fmt.Sprintf("%d-%s", time.Now().Unix(), header.Filename)
	destPath := filepath.Join(uploadsDir, id)

	// Create destination file
	dst, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy uploaded file
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	log.Info().Str("filename", header.Filename).Str("id", id).Msg("File uploaded successfully")

	response := map[string]interface{}{
		"id":       id,
		"filename": header.Filename,
		"size":     header.Size,
		"status":   "uploaded",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleAnalyzePCAP starts analysis of an uploaded PCAP file
func (h *Handler) HandleAnalyzePCAP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	// Check if file exists
	uploadsDir := "ttmp/2025-07-13/pcap/uploads"
	filePath := filepath.Join(uploadsDir, id)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Move file to captures directory for analysis
	capturesDir := "ttmp/2025-07-13/pcap/captures"
	os.MkdirAll(capturesDir, 0755)

	capturePath := filepath.Join(capturesDir, id)
	if err := os.Rename(filePath, capturePath); err != nil {
		log.Error().Err(err).Msg("Failed to move file to captures directory")
		// Continue with analysis even if move fails
		capturePath = filePath
	}

	// Start analysis in background
	go func() {
		sessionPrefix := strings.TrimSuffix(id, filepath.Ext(id))
		scriptPath := "ttmp/2025-07-13/pcap/analyze-ppa-captures.sh"
		cmd := exec.Command("bash", scriptPath, sessionPrefix)
		cmd.Dir = "."

		log.Info().Str("session", sessionPrefix).Msg("Starting background analysis")

		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Error().Err(err).Str("output", string(output)).Str("session", sessionPrefix).Msg("Analysis failed")
		} else {
			log.Info().Str("session", sessionPrefix).Msg("Analysis completed")
		}
	}()

	response := map[string]interface{}{
		"status":  "analyzing",
		"message": "Analysis started",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandlePCAPStatus returns the status of a PCAP analysis
func (h *Handler) HandlePCAPStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	sessionPrefix := strings.TrimSuffix(id, filepath.Ext(id))

	// Check if analysis files exist
	analysisDir := "ttmp/2025-07-13/pcap/analysis"
	files, err := os.ReadDir(analysisDir)
	if err != nil {
		response := map[string]interface{}{
			"status":   "error",
			"progress": 0,
			"message":  "Analysis directory not found",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	hasResults := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), sessionPrefix) {
			hasResults = true
			break
		}
	}

	var response map[string]interface{}
	if hasResults {
		response = map[string]interface{}{
			"status":   "completed",
			"progress": 100,
			"message":  "Analysis completed",
		}
	} else {
		response = map[string]interface{}{
			"status":   "analyzing",
			"progress": 50,
			"message":  "Analysis in progress",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleListPCAPs lists uploaded PCAP files
func (h *Handler) HandleListPCAPs(w http.ResponseWriter, r *http.Request) {
	var allFiles []map[string]interface{}

	// Check uploads directory
	uploadsDir := "ttmp/2025-07-13/pcap/uploads"
	if files, err := os.ReadDir(uploadsDir); err == nil {
		for _, file := range files {
			if info, err := file.Info(); err == nil {
				allFiles = append(allFiles, map[string]interface{}{
					"id":         file.Name(),
					"name":       file.Name(),
					"size":       info.Size(),
					"uploadDate": info.ModTime(),
					"status":     "uploaded",
				})
			}
		}
	}

	// Check captures directory
	capturesDir := "ttmp/2025-07-13/pcap/captures"
	if files, err := os.ReadDir(capturesDir); err == nil {
		for _, file := range files {
			if info, err := file.Info(); err == nil {
				// Check if analysis exists
				sessionPrefix := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
				status := "uploaded"

				analysisDir := "ttmp/2025-07-13/pcap/analysis"
				if analysisFiles, err := os.ReadDir(analysisDir); err == nil {
					for _, analysisFile := range analysisFiles {
						if strings.HasPrefix(analysisFile.Name(), sessionPrefix) {
							status = "analyzed"
							break
						}
					}
				}

				allFiles = append(allFiles, map[string]interface{}{
					"id":         file.Name(),
					"name":       file.Name(),
					"size":       info.Size(),
					"uploadDate": info.ModTime(),
					"status":     status,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allFiles)
}

// HandleGetAnalysisResult returns analysis results for a specific analysis
func (h *Handler) HandleGetAnalysisResult(w http.ResponseWriter, r *http.Request) {
	// For now, redirect to the legacy handler
	h.HandleGetAnalysis(w, r)
}

// HandleListAnalysisResults lists all available analysis results
func (h *Handler) HandleListAnalysisResults(w http.ResponseWriter, r *http.Request) {
	analysisDir := "ttmp/2025-07-13/pcap/analysis"
	files, err := os.ReadDir(analysisDir)
	if err != nil {
		http.Error(w, "Failed to read analysis directory", http.StatusInternalServerError)
		return
	}

	// Group files by session
	sessions := make(map[string][]types.AnalysisFile)
	for _, file := range files {
		// Extract session prefix from filename
		name := file.Name()
		parts := strings.Split(name, "-")
		if len(parts) >= 2 {
			session := strings.Join(parts[:2], "-")

			if info, err := file.Info(); err == nil {
				sessions[session] = append(sessions[session], types.AnalysisFile{
					Name:         name,
					Type:         getFileType(name),
					Size:         info.Size(),
					ModifiedTime: info.ModTime(),
					Path:         filepath.Join(analysisDir, name),
				})
			}
		}
	}

	var results []types.AnalysisResult
	for session, files := range sessions {
		results = append(results, types.AnalysisResult{
			Session: session,
			Files:   files,
		})
	}

	// Sort by session name
	sort.Slice(results, func(i, j int) bool {
		return results[i].Session > results[j].Session // Newer first
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// HandleListCaptures lists available PCAP files
func (h *Handler) HandleListCaptures(w http.ResponseWriter, r *http.Request) {
	capturesDir := "ttmp/2025-07-13/pcap/captures"
	files, err := os.ReadDir(capturesDir)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read captures directory")
		http.Error(w, "Failed to read captures directory", http.StatusInternalServerError)
		return
	}

	var captures []types.CaptureFile
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pcap") {
			info, err := file.Info()
			if err != nil {
				continue
			}

			captures = append(captures, types.CaptureFile{
				Name:         file.Name(),
				Size:         info.Size(),
				ModifiedTime: info.ModTime(),
				Path:         filepath.Join(capturesDir, file.Name()),
			})
		}
	}

	// Sort by modification time, newest first
	sort.Slice(captures, func(i, j int) bool {
		return captures[i].ModifiedTime.After(captures[j].ModifiedTime)
	})

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(captures)
}

// HandleAnalyze triggers analysis of a PCAP file
func (h *Handler) HandleAnalyze(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	if filename == "" {
		http.Error(w, "Filename is required", http.StatusBadRequest)
		return
	}

	// Extract session prefix from filename (e.g., "preset-switching-132325" from "preset-switching-132325.pcap")
	sessionPrefix := strings.TrimSuffix(filename, ".pcap")

	// Run analysis script
	scriptPath := "ttmp/2025-07-13/pcap/analyze-ppa-captures.sh"
	cmd := exec.Command("bash", scriptPath, sessionPrefix)
	cmd.Dir = "."

	log.Info().Str("session", sessionPrefix).Msg("Starting analysis")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error().Err(err).Str("output", string(output)).Msg("Analysis failed")
		http.Error(w, fmt.Sprintf("Analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := types.AnalysisResponse{
		Session:   sessionPrefix,
		Status:    "completed",
		Output:    string(output),
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// HandleGetAnalysis gets analysis results for a session
func (h *Handler) HandleGetAnalysis(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session := vars["session"]

	if session == "" {
		http.Error(w, "Session is required", http.StatusBadRequest)
		return
	}

	analysisDir := "ttmp/2025-07-13/pcap/analysis"
	files, err := os.ReadDir(analysisDir)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read analysis directory")
		http.Error(w, "Failed to read analysis directory", http.StatusInternalServerError)
		return
	}

	var analysisFiles []types.AnalysisFile
	for _, file := range files {
		if strings.HasPrefix(file.Name(), session) {
			info, err := file.Info()
			if err != nil {
				continue
			}

			analysisFiles = append(analysisFiles, types.AnalysisFile{
				Name:         file.Name(),
				Type:         getFileType(file.Name()),
				Size:         info.Size(),
				ModifiedTime: info.ModTime(),
				Path:         filepath.Join(analysisDir, file.Name()),
			})
		}
	}

	// Sort by name for consistent ordering
	sort.Slice(analysisFiles, func(i, j int) bool {
		return analysisFiles[i].Name < analysisFiles[j].Name
	})

	response := types.AnalysisResult{
		Session: session,
		Files:   analysisFiles,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// HandleServeFile serves analysis files (markdown, JSON, text)
func (h *Handler) HandleServeFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filePath := vars["path"]

	// Security: prevent path traversal
	if strings.Contains(filePath, "..") {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Only serve files from analysis directory
	fullPath := filepath.Join("ttmp/2025-07-13/pcap/analysis", filePath)

	// Check if file exists and is within allowed directory
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, "ttmp/2025-07-13/pcap/analysis") {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		log.Error().Err(err).Str("path", cleanPath).Msg("Failed to open file")
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Set appropriate content type
	ext := filepath.Ext(filePath)
	switch ext {
	case ".json":
		w.Header().Set("Content-Type", "application/json")
	case ".md":
		w.Header().Set("Content-Type", "text/markdown")
	case ".txt":
		w.Header().Set("Content-Type", "text/plain")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	io.Copy(w, file)
}

// HandleSearch searches through analysis documents
func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	analysisDir := "ttmp/2025-07-13/pcap/analysis"
	files, err := os.ReadDir(analysisDir)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read analysis directory")
		http.Error(w, "Failed to read analysis directory", http.StatusInternalServerError)
		return
	}

	var results []types.SearchResult
	queryLower := strings.ToLower(query)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(analysisDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}

		contentStr := string(content)
		contentLower := strings.ToLower(contentStr)

		if strings.Contains(contentLower, queryLower) {
			lines := strings.Split(contentStr, "\n")
			var matchingLines []types.SearchMatch

			for i, line := range lines {
				if strings.Contains(strings.ToLower(line), queryLower) {
					matchingLines = append(matchingLines, types.SearchMatch{
						LineNumber: i + 1,
						Line:       line,
						Context:    getLineContext(lines, i, 2),
					})

					// Limit matches per file
					if len(matchingLines) >= 10 {
						break
					}
				}
			}

			if len(matchingLines) > 0 {
				results = append(results, types.SearchResult{
					File:    file.Name(),
					Matches: matchingLines,
				})
			}
		}
	}

	response := types.SearchResponse{
		Query:   query,
		Results: results,
		Count:   len(results),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(response)
}

// HandleGetPackets gets packet data for visualization
func (h *Handler) HandleGetPackets(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	session := vars["session"]

	if session == "" {
		http.Error(w, "Session is required", http.StatusBadRequest)
		return
	}

	// Look for JSON file with packet data
	jsonFile := filepath.Join("ttmp/2025-07-13/pcap/analysis", session+".json")

	file, err := os.Open(jsonFile)
	if err != nil {
		log.Error().Err(err).Str("file", jsonFile).Msg("Failed to open packet JSON file")
		http.Error(w, "Packet data not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	io.Copy(w, file)
}

// Helper functions

func getFileType(filename string) string {
	if strings.Contains(filename, "summary-report") {
		return "summary"
	}
	if strings.Contains(filename, "message-types") {
		return "message_types"
	}
	if strings.Contains(filename, "status-codes") {
		return "status_codes"
	}
	if strings.Contains(filename, "sequences") {
		return "sequences"
	}
	if strings.Contains(filename, "payloads") {
		return "payloads"
	}
	if strings.Contains(filename, "unknown") {
		return "unknown"
	}
	if strings.Contains(filename, "livecmd") {
		return "livecmd"
	}
	if strings.HasSuffix(filename, ".json") {
		return "packets"
	}
	return "other"
}

func getLineContext(lines []string, index, contextSize int) []string {
	start := index - contextSize
	if start < 0 {
		start = 0
	}

	end := index + contextSize + 1
	if end > len(lines) {
		end = len(lines)
	}

	return lines[start:end]
}
