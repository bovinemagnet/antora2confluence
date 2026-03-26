package renderer

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bovinemagnet/antora2confluence/internal/model"
)

// Renderer converts AsciiDoc pages to Confluence storage format
// via the asciidoctor CLI and HTML transformation.
type Renderer struct {
	pageTitles    PageTitleMap
	backendConfig BackendConfig
	mermaidMode   string
}

// New creates a new Renderer.
func New(cfg BackendConfig, mermaidMode string) *Renderer {
	return &Renderer{backendConfig: cfg, mermaidMode: mermaidMode}
}

// Render converts a single Page into a RenderedPage by running
// asciidoctor and transforming the HTML output.
func (r *Renderer) Render(page model.Page) (*model.RenderedPage, error) {
	htmlContent, err := ConvertToHTML(page.AbsPath, r.backendConfig)
	if err != nil {
		return nil, fmt.Errorf("rendering %s: %w", page.RelPath, err)
	}

	body, err := TransformToConfluence(htmlContent, r.pageTitles, r.mermaidMode)
	if err != nil {
		return nil, fmt.Errorf("transforming %s: %w", page.RelPath, err)
	}

	title := page.Title
	fingerprint := computeFingerprint(title, body)

	return &model.RenderedPage{
		SourcePage:  page,
		Title:       title,
		Body:        body,
		Fingerprint: fingerprint,
		Includes:    page.Includes,
		Images:      page.Images,
		XRefs:       page.XRefs,
	}, nil
}

// RenderAll renders a slice of pages. Pages that fail to render are
// collected as errors rather than stopping the entire batch.
func (r *Renderer) RenderAll(pages []model.Page) ([]model.RenderedPage, []error) {
	// Build page title map for xref resolution
	r.pageTitles = make(PageTitleMap, len(pages))
	for _, page := range pages {
		baseName := strings.TrimSuffix(page.RelPath, ".adoc")
		r.pageTitles[baseName] = page.Title
	}

	var rendered []model.RenderedPage
	var errs []error

	for _, page := range pages {
		rp, err := r.Render(page)
		if err != nil {
			slog.Warn("Failed to render page", "page", page.PageKey, "error", err)
			errs = append(errs, err)
			continue
		}
		rendered = append(rendered, *rp)
	}

	return rendered, errs
}

func computeFingerprint(title, body string) string {
	h := sha256.New()
	h.Write([]byte(title))
	h.Write([]byte("\x00"))
	h.Write([]byte(body))
	return fmt.Sprintf("%x", h.Sum(nil))
}
