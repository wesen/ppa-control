package templates

import (
	"context"
	"net/http"
	"ppa-control/cmd/ppa-web/server"

	"github.com/a-h/templ"
)

// TemplateProvider implements server.TemplateProvider interface
type TemplateProvider struct{}

// NewTemplateProvider creates a new TemplateProvider instance
func NewTemplateProvider() server.TemplateProvider {
	return &TemplateProvider{}
}

// Index returns the index page template
func (p *TemplateProvider) Index(state server.AppState) server.TemplateRenderer {
	return templateRenderer{Index(state)}
}

// StatusBar returns the status bar template
func (p *TemplateProvider) StatusBar(state server.AppState) server.TemplateRenderer {
	return templateRenderer{StatusBar(state)}
}

// IPForm returns the IP form template
func (p *TemplateProvider) IPForm(state server.AppState) server.TemplateRenderer {
	return templateRenderer{IPForm(state)}
}

// LogWindow returns the log window template
func (p *TemplateProvider) LogWindow(state server.AppState) server.TemplateRenderer {
	return templateRenderer{LogWindow(state)}
}

// templateRenderer wraps a templ.Component to implement server.TemplateRenderer
type templateRenderer struct {
	component templ.Component
}

// Render implements server.TemplateRenderer
func (t templateRenderer) Render(ctx context.Context, w http.ResponseWriter) error {
	return t.component.Render(ctx, w)
}
