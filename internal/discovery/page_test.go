package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePage_ExtractsTitle(t *testing.T) {
	page, err := ParsePage(
		testdataPath("valid-repo/modules/ROOT/pages/index.adoc"),
		"index.adoc",
		"site/billing/1.0/ROOT/index.adoc",
	)
	require.NoError(t, err)
	assert.Equal(t, "Overview", page.Title)
}

func TestParsePage_ExtractsExplicitKey(t *testing.T) {
	page, err := ParsePage(
		testdataPath("valid-repo/modules/ROOT/pages/index.adoc"),
		"index.adoc",
		"site/billing/1.0/ROOT/index.adoc",
	)
	require.NoError(t, err)
	assert.Equal(t, "billing-overview", page.ExplicitKey)
}

func TestParsePage_ExtractsImages(t *testing.T) {
	page, err := ParsePage(
		testdataPath("valid-repo/modules/ROOT/pages/index.adoc"),
		"index.adoc",
		"site/billing/1.0/ROOT/index.adoc",
	)
	require.NoError(t, err)
	assert.Contains(t, page.Images, "logo.png")
}

func TestParsePage_ExtractsXRefs(t *testing.T) {
	page, err := ParsePage(
		testdataPath("valid-repo/modules/ROOT/pages/index.adoc"),
		"index.adoc",
		"site/billing/1.0/ROOT/index.adoc",
	)
	require.NoError(t, err)
	assert.Contains(t, page.XRefs, "getting-started.adoc")
}

func TestParsePage_ExtractsIncludes(t *testing.T) {
	page, err := ParsePage(
		testdataPath("valid-repo/modules/ROOT/pages/getting-started.adoc"),
		"getting-started.adoc",
		"site/billing/1.0/ROOT/getting-started.adoc",
	)
	require.NoError(t, err)
	assert.Contains(t, page.Includes, "partial$setup-steps.adoc")
}

func TestParsePage_SetsRelAndAbsPaths(t *testing.T) {
	absPath := testdataPath("valid-repo/modules/ROOT/pages/index.adoc")
	page, err := ParsePage(absPath, "index.adoc", "site/billing/1.0/ROOT/index.adoc")
	require.NoError(t, err)
	assert.Equal(t, "index.adoc", page.RelPath)
	assert.Equal(t, absPath, page.AbsPath)
}

func TestParsePage_NoTitle_UsesFilename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no-title.adoc")
	os.WriteFile(path, []byte("Just some content without a title.\n"), 0644)

	page, err := ParsePage(path, "no-title.adoc", "site/comp/1.0/mod/no-title.adoc")
	require.NoError(t, err)
	assert.Equal(t, "no-title", page.Title)
}
