package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseNav_ValidNavFile_ReturnsEntries(t *testing.T) {
	entries, err := ParseNav(testdataPath("valid-repo/modules/ROOT/nav.adoc"))
	require.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, "Overview", entries[0].Title)
	assert.Equal(t, "index.adoc", entries[0].PageRef)
	assert.Equal(t, 0, entries[0].Order)
	assert.Equal(t, "Getting Started", entries[1].Title)
	assert.Equal(t, "getting-started.adoc", entries[1].PageRef)
	assert.Equal(t, 1, entries[1].Order)
}

func TestParseNav_NonexistentFile_ReturnsOsError(t *testing.T) {
	_, err := ParseNav("/nonexistent/nav.adoc")
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestParseNav_EmptyFile_ReturnsEmptySlice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nav.adoc")
	os.WriteFile(path, []byte(""), 0644)

	entries, err := ParseNav(path)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestParseNav_NestedEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nav.adoc")
	content := "* xref:parent.adoc[Parent]\n** xref:child.adoc[Child]\n"
	os.WriteFile(path, []byte(content), 0644)

	entries, err := ParseNav(path)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "Parent", entries[0].Title)
	require.Len(t, entries[0].Children, 1)
	assert.Equal(t, "Child", entries[0].Children[0].Title)
	assert.Equal(t, "child.adoc", entries[0].Children[0].PageRef)
}

func TestParseNav_PlainTextEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nav.adoc")
	content := "* Getting Started\n** xref:setup.adoc[Setup]\n"
	os.WriteFile(path, []byte(content), 0644)

	entries, err := ParseNav(path)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "Getting Started", entries[0].Title)
	assert.Equal(t, "", entries[0].PageRef)
	require.Len(t, entries[0].Children, 1)
	assert.Equal(t, "Setup", entries[0].Children[0].Title)
}
