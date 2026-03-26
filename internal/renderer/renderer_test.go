package renderer

import (
	"os/exec"
	"testing"

	"github.com/bovinemagnet/antora2confluence/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRender_ValidPage_ReturnsRenderedPage(t *testing.T) {
	if _, err := exec.LookPath("asciidoctor"); err != nil {
		t.Skip("asciidoctor not found on PATH, skipping")
	}

	r := New(DefaultBackendConfig(), "passthrough")
	page := model.Page{
		RelPath: "index.adoc",
		AbsPath: testdataPath("valid-repo/modules/ROOT/pages/index.adoc"),
		Title:   "Overview",
		PageKey: "test-site/billing/1.0/ROOT/index.adoc",
	}

	rendered, err := r.Render(page)
	require.NoError(t, err)
	assert.Equal(t, "Overview", rendered.Title)
	assert.NotEmpty(t, rendered.Body)
	assert.NotEmpty(t, rendered.Fingerprint)
	assert.Contains(t, rendered.Body, "billing service overview")
}

func TestRender_PageWithCodeBlock_ContainsCodeMacro(t *testing.T) {
	if _, err := exec.LookPath("asciidoctor"); err != nil {
		t.Skip("asciidoctor not found on PATH, skipping")
	}

	r := New(DefaultBackendConfig(), "passthrough")
	page := model.Page{
		RelPath: "authentication.adoc",
		AbsPath: testdataPath("valid-repo/modules/api/pages/authentication.adoc"),
		Title:   "Authentication",
		PageKey: "test-site/billing/1.0/api/authentication.adoc",
	}

	rendered, err := r.Render(page)
	require.NoError(t, err)
	assert.Contains(t, rendered.Body, `ac:name="code"`)
	assert.Contains(t, rendered.Body, "bash")
}

func TestRender_PageWithTable_ContainsTableMarkup(t *testing.T) {
	if _, err := exec.LookPath("asciidoctor"); err != nil {
		t.Skip("asciidoctor not found on PATH, skipping")
	}

	r := New(DefaultBackendConfig(), "passthrough")
	page := model.Page{
		RelPath: "errors.adoc",
		AbsPath: testdataPath("valid-repo/modules/api/pages/errors.adoc"),
		Title:   "Error Handling",
		PageKey: "test-site/billing/1.0/api/errors.adoc",
	}

	rendered, err := r.Render(page)
	require.NoError(t, err)
	assert.Contains(t, rendered.Body, "<table>")
	assert.Contains(t, rendered.Body, "400")
}

func TestRender_SameContent_SameFingerprint(t *testing.T) {
	if _, err := exec.LookPath("asciidoctor"); err != nil {
		t.Skip("asciidoctor not found on PATH, skipping")
	}

	r := New(DefaultBackendConfig(), "passthrough")
	page := model.Page{
		RelPath: "index.adoc",
		AbsPath: testdataPath("valid-repo/modules/ROOT/pages/index.adoc"),
		Title:   "Overview",
		PageKey: "test-site/billing/1.0/ROOT/index.adoc",
	}

	r1, err := r.Render(page)
	require.NoError(t, err)
	r2, err := r.Render(page)
	require.NoError(t, err)

	assert.Equal(t, r1.Fingerprint, r2.Fingerprint)
}

func TestRender_DifferentPages_DifferentFingerprints(t *testing.T) {
	if _, err := exec.LookPath("asciidoctor"); err != nil {
		t.Skip("asciidoctor not found on PATH, skipping")
	}

	r := New(DefaultBackendConfig(), "passthrough")
	page1 := model.Page{
		RelPath: "index.adoc",
		AbsPath: testdataPath("valid-repo/modules/ROOT/pages/index.adoc"),
		Title:   "Overview",
		PageKey: "test-site/billing/1.0/ROOT/index.adoc",
	}
	page2 := model.Page{
		RelPath: "authentication.adoc",
		AbsPath: testdataPath("valid-repo/modules/api/pages/authentication.adoc"),
		Title:   "Authentication",
		PageKey: "test-site/billing/1.0/api/authentication.adoc",
	}

	r1, err := r.Render(page1)
	require.NoError(t, err)
	r2, err := r.Render(page2)
	require.NoError(t, err)

	assert.NotEqual(t, r1.Fingerprint, r2.Fingerprint)
}

func TestRenderAll_ValidPages_ReturnsAllRendered(t *testing.T) {
	if _, err := exec.LookPath("asciidoctor"); err != nil {
		t.Skip("asciidoctor not found on PATH, skipping")
	}

	r := New(DefaultBackendConfig(), "passthrough")
	pages := []model.Page{
		{
			RelPath: "index.adoc",
			AbsPath: testdataPath("valid-repo/modules/ROOT/pages/index.adoc"),
			Title:   "Overview",
			PageKey: "test-site/billing/1.0/ROOT/index.adoc",
		},
		{
			RelPath: "authentication.adoc",
			AbsPath: testdataPath("valid-repo/modules/api/pages/authentication.adoc"),
			Title:   "Authentication",
			PageKey: "test-site/billing/1.0/api/authentication.adoc",
		},
	}

	rendered, errs := r.RenderAll(pages)
	assert.Empty(t, errs)
	assert.Len(t, rendered, 2)
}
