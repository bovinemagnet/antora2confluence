package discovery

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bovinemagnet/antora2confluence/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataPath(name string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata", name)
}

func TestScan_ValidRepo_ReturnsContentSource(t *testing.T) {
	source, err := Scan(testdataPath("valid-repo"), "test-site")
	require.NoError(t, err)

	assert.Equal(t, "test-site", source.SiteKey)
	assert.Len(t, source.Components, 1)

	comp := source.Components[0]
	assert.Equal(t, "billing", comp.Name)
	assert.Len(t, comp.Versions, 1)

	ver := comp.Versions[0]
	assert.Equal(t, "1.0", ver.Name)
	assert.Len(t, ver.Modules, 2)
}

func TestScan_ValidRepo_DiscoversModulesAndPages(t *testing.T) {
	source, err := Scan(testdataPath("valid-repo"), "test-site")
	require.NoError(t, err)

	ver := source.Components[0].Versions[0]

	var rootMod, apiMod *model.Module
	for i := range ver.Modules {
		switch ver.Modules[i].Name {
		case "ROOT":
			rootMod = &ver.Modules[i]
		case "api":
			apiMod = &ver.Modules[i]
		}
	}

	require.NotNil(t, rootMod, "ROOT module should exist")
	require.NotNil(t, apiMod, "api module should exist")

	assert.Len(t, rootMod.Pages, 2)
	assert.Len(t, apiMod.Pages, 2)
}

func TestScan_ValidRepo_SetsPageKeys(t *testing.T) {
	source, err := Scan(testdataPath("valid-repo"), "test-site")
	require.NoError(t, err)

	pages := source.AllPages()
	keys := make(map[string]bool)
	for _, p := range pages {
		keys[p.PageKey] = true
	}

	assert.True(t, keys["test-site/billing/1.0/ROOT/index.adoc"])
	assert.True(t, keys["test-site/billing/1.0/ROOT/getting-started.adoc"])
	assert.True(t, keys["test-site/billing/1.0/api/authentication.adoc"])
	assert.True(t, keys["test-site/billing/1.0/api/errors.adoc"])
}

func TestScan_InvalidRepo_MissingName_ReturnsError(t *testing.T) {
	_, err := Scan(testdataPath("invalid-repo"), "test-site")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name")
}

func TestScan_NonexistentPath_ReturnsError(t *testing.T) {
	_, err := Scan("/nonexistent/path", "test-site")
	require.Error(t, err)
}
