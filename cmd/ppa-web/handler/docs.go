package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// DocumentMetadata contains metadata for a document
type DocumentMetadata struct {
	Path         string    `json:"path"`
	RelativePath string    `json:"relativePath"`
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"modTime"`
	Type         string    `json:"type"`
	Title        string    `json:"title,omitempty"`
	Summary      string    `json:"summary,omitempty"`
}

// DocumentSearchResult represents a search result for documents
type DocumentSearchResult struct {
	Document DocumentMetadata      `json:"document"`
	Matches  []DocumentSearchMatch `json:"matches"`
	Score    int                   `json:"score"`
}

// DocumentSearchMatch represents a match within a document
type DocumentSearchMatch struct {
	LineNumber int    `json:"lineNumber"`
	Line       string `json:"line"`
	Context    string `json:"context"`
}

// DocumentService handles document operations
type DocumentService struct {
	basePath string
	index    map[string]DocumentMetadata
}

// NewDocumentService creates a new document service
func NewDocumentService(basePath string) *DocumentService {
	return &DocumentService{
		basePath: basePath,
		index:    make(map[string]DocumentMetadata),
	}
}

// BuildIndex scans the filesystem and builds the document index
func (ds *DocumentService) BuildIndex() error {
	ds.index = make(map[string]DocumentMetadata)

	// Define paths to index
	pathsToIndex := []string{
		"doc",
		"ttmp",
		"README.md",
		"changelog.md",
		"tutorial-protocol.md",
	}

	for _, path := range pathsToIndex {
		fullPath := filepath.Join(ds.basePath, path)
		if err := ds.indexPath(fullPath, path); err != nil {
			log.Warn().Err(err).Str("path", path).Msg("Failed to index path")
		}
	}

	log.Info().Int("documents", len(ds.index)).Msg("Document index built")
	return nil
}

// indexPath recursively indexes documents in a path
func (ds *DocumentService) indexPath(fullPath, relativePath string) error {
	info, err := os.Stat(fullPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			entryFullPath := filepath.Join(fullPath, entry.Name())
			entryRelativePath := filepath.Join(relativePath, entry.Name())

			if err := ds.indexPath(entryFullPath, entryRelativePath); err != nil {
				log.Warn().Err(err).Str("path", entryRelativePath).Msg("Failed to index entry")
			}
		}
	} else {
		// Only index markdown and text files
		ext := strings.ToLower(filepath.Ext(fullPath))
		if ext == ".md" || ext == ".txt" || ext == ".log" {
			metadata := DocumentMetadata{
				Path:         fullPath,
				RelativePath: relativePath,
				Name:         info.Name(),
				Size:         info.Size(),
				ModTime:      info.ModTime(),
				Type:         ext,
			}

			// Extract title and summary for markdown files
			if ext == ".md" {
				if title, summary, err := ds.extractMarkdownMetadata(fullPath); err == nil {
					metadata.Title = title
					metadata.Summary = summary
				}
			}

			ds.index[relativePath] = metadata
		}
	}

	return nil
}

// extractMarkdownMetadata extracts title and summary from markdown content
func (ds *DocumentService) extractMarkdownMetadata(path string) (title, summary string, err error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	lines := strings.Split(string(content), "\n")

	// Look for first heading as title
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			title = strings.TrimPrefix(line, "# ")
			break
		}
	}

	// Create summary from first few non-empty, non-heading lines
	summaryLines := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		summaryLines = append(summaryLines, line)
		if len(summaryLines) >= 3 {
			break
		}
	}
	summary = strings.Join(summaryLines, " ")
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}

	return title, summary, nil
}

// Search performs full-text search across indexed documents
func (ds *DocumentService) Search(query string, fileTypes []string, maxResults int) ([]DocumentSearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	query = strings.ToLower(query)
	results := []DocumentSearchResult{}

	for _, doc := range ds.index {
		// Filter by file type if specified
		if len(fileTypes) > 0 {
			found := false
			for _, ft := range fileTypes {
				if doc.Type == ft {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		matches, score := ds.searchInDocument(doc, query)
		if score > 0 {
			results = append(results, DocumentSearchResult{
				Document: doc,
				Matches:  matches,
				Score:    score,
			})
		}
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if maxResults > 0 && len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}

// searchInDocument searches for query in a specific document
func (ds *DocumentService) searchInDocument(doc DocumentMetadata, query string) ([]DocumentSearchMatch, int) {
	content, err := os.ReadFile(doc.Path)
	if err != nil {
		return nil, 0
	}

	lines := strings.Split(string(content), "\n")
	matches := []DocumentSearchMatch{}
	score := 0

	for i, line := range lines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, query) {
			// Count occurrences for scoring
			score += strings.Count(lowerLine, query)

			// Create context (surrounding lines)
			contextStart := i - 2
			contextEnd := i + 3
			if contextStart < 0 {
				contextStart = 0
			}
			if contextEnd > len(lines) {
				contextEnd = len(lines)
			}

			context := strings.Join(lines[contextStart:contextEnd], "\n")

			matches = append(matches, DocumentSearchMatch{
				LineNumber: i + 1,
				Line:       line,
				Context:    context,
			})
		}
	}

	return matches, score
}

// GetDocuments returns a list of all indexed documents
func (ds *DocumentService) GetDocuments() []DocumentMetadata {
	docs := make([]DocumentMetadata, 0, len(ds.index))
	for _, doc := range ds.index {
		docs = append(docs, doc)
	}

	// Sort by modification time (newest first)
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].ModTime.After(docs[j].ModTime)
	})

	return docs
}

// GetDocument returns a specific document by path
func (ds *DocumentService) GetDocument(path string) (DocumentMetadata, bool) {
	doc, exists := ds.index[path]
	return doc, exists
}

// convertMarkdownToHTML converts markdown content to HTML
func convertMarkdownToHTML(content []byte) ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Table,
			extension.Strikethrough,
			extension.Linkify,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf strings.Builder
	if err := md.Convert(content, &buf); err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}

// HandleMarkdownRender converts markdown to HTML
func (h *Handler) HandleMarkdownRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	html, err := convertMarkdownToHTML([]byte(req.Content))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"html": string(html),
	})
}

// HandleDocumentSearch handles document search requests
func (h *Handler) HandleDocumentSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	fileTypes := r.URL.Query()["type"]

	if h.docService == nil {
		http.Error(w, "Document service not available", http.StatusServiceUnavailable)
		return
	}

	results, err := h.docService.Search(query, fileTypes, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// HandleDocumentList handles document listing requests
func (h *Handler) HandleDocumentList(w http.ResponseWriter, r *http.Request) {
	if h.docService == nil {
		http.Error(w, "Document service not available", http.StatusServiceUnavailable)
		return
	}

	docs := h.docService.GetDocuments()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

// HandleDocumentView handles viewing a specific document
func (h *Handler) HandleDocumentView(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Path parameter required", http.StatusBadRequest)
		return
	}

	if h.docService == nil {
		http.Error(w, "Document service not available", http.StatusServiceUnavailable)
		return
	}

	doc, exists := h.docService.GetDocument(path)
	if !exists {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	content, err := os.ReadFile(doc.Path)
	if err != nil {
		http.Error(w, "Failed to read document", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"document": doc,
		"content":  string(content),
	}

	// Convert markdown to HTML if it's a markdown file
	if doc.Type == ".md" {
		html, err := convertMarkdownToHTML(content)
		if err == nil {
			response["html"] = string(html)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
