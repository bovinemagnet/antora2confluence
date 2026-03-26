package renderer

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata", name)
}

func skipIfNoAsciidoctor(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("asciidoctor"); err != nil {
		t.Skip("asciidoctor not found on PATH, skipping")
	}
}

func TestConvertToHTML_ValidFile_ReturnsHTML(t *testing.T) {
	skipIfNoAsciidoctor(t)
	html, err := ConvertToHTML(testdataPath("valid-repo/modules/ROOT/pages/index.adoc"), DefaultBackendConfig())
	require.NoError(t, err)
	assert.Contains(t, html, "billing service overview")
}

func TestConvertToHTML_ValidFile_ContainsHTMLTags(t *testing.T) {
	skipIfNoAsciidoctor(t)
	html, err := ConvertToHTML(testdataPath("valid-repo/modules/ROOT/pages/index.adoc"), DefaultBackendConfig())
	require.NoError(t, err)
	assert.Contains(t, html, "<")
	assert.Contains(t, html, ">")
}

func TestConvertToHTML_FileWithCodeBlock_ContainsPreTag(t *testing.T) {
	skipIfNoAsciidoctor(t)
	html, err := ConvertToHTML(testdataPath("valid-repo/modules/api/pages/authentication.adoc"), DefaultBackendConfig())
	require.NoError(t, err)
	assert.Contains(t, html, "<pre")
}

func TestConvertToHTML_FileWithTable_ContainsTableTag(t *testing.T) {
	skipIfNoAsciidoctor(t)
	html, err := ConvertToHTML(testdataPath("valid-repo/modules/api/pages/errors.adoc"), DefaultBackendConfig())
	require.NoError(t, err)
	assert.Contains(t, html, "<table")
}

func TestConvertToHTML_NonexistentFile_ReturnsError(t *testing.T) {
	skipIfNoAsciidoctor(t)
	_, err := ConvertToHTML("/nonexistent/file.adoc", DefaultBackendConfig())
	require.Error(t, err)
}

func TestBuildAsciidoctorArgs_WithExtensions(t *testing.T) {
	cfg := BackendConfig{
		Backend:    "local",
		Extensions: []string{"asciidoctor-diagram", "asciidoctor-kroki"},
	}
	args := buildAsciidoctorArgs("/tmp/test.adoc", cfg)
	assert.Contains(t, args, "-r")
	assert.Contains(t, args, "asciidoctor-diagram")
	assert.Contains(t, args, "asciidoctor-kroki")
}

func TestBuildAsciidoctorArgs_NoExtensions(t *testing.T) {
	cfg := DefaultBackendConfig()
	args := buildAsciidoctorArgs("/tmp/test.adoc", cfg)
	assert.NotContains(t, args, "-r")
	assert.Contains(t, args, "/tmp/test.adoc")
}
