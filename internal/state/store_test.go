package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_NewEmpty(t *testing.T) {
	s := New("/tmp/test-state.json")
	assert.Empty(t, s.Entries())
}

func TestStore_Upsert_And_Lookup(t *testing.T) {
	s := New("/tmp/test-state.json")
	s.Upsert(Entry{PageKey: "site/billing/1.0/ROOT/index.adoc", ConfluenceID: "123", Title: "Overview", Fingerprint: "abc123", ParentID: "100", Version: 1, PublishedAt: time.Now()})
	got, ok := s.Lookup("site/billing/1.0/ROOT/index.adoc")
	require.True(t, ok)
	assert.Equal(t, "123", got.ConfluenceID)
	assert.Equal(t, "abc123", got.Fingerprint)
}

func TestStore_Lookup_NotFound(t *testing.T) {
	s := New("/tmp/test-state.json")
	_, ok := s.Lookup("nonexistent")
	assert.False(t, ok)
}

func TestStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	s := New(path)
	s.Upsert(Entry{PageKey: "site/billing/1.0/ROOT/index.adoc", ConfluenceID: "123", Fingerprint: "abc123", PublishedAt: time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC)})
	s.Upsert(Entry{PageKey: "site/billing/1.0/api/auth.adoc", ConfluenceID: "456", Fingerprint: "def456", PublishedAt: time.Date(2026, 3, 27, 0, 0, 0, 0, time.UTC)})
	require.NoError(t, s.Save())

	s2 := New(path)
	require.NoError(t, s2.Load())
	assert.Len(t, s2.Entries(), 2)
	got, ok := s2.Lookup("site/billing/1.0/ROOT/index.adoc")
	require.True(t, ok)
	assert.Equal(t, "123", got.ConfluenceID)
}

func TestStore_Load_NonexistentFile_ReturnsNoError(t *testing.T) {
	s := New("/nonexistent/path/state.json")
	require.NoError(t, s.Load())
	assert.Empty(t, s.Entries())
}

func TestStore_Upsert_OverwritesExisting(t *testing.T) {
	s := New("/tmp/test-state.json")
	s.Upsert(Entry{PageKey: "key1", Fingerprint: "old"})
	s.Upsert(Entry{PageKey: "key1", Fingerprint: "new"})
	got, ok := s.Lookup("key1")
	require.True(t, ok)
	assert.Equal(t, "new", got.Fingerprint)
	assert.Len(t, s.Entries(), 1)
}

func TestStore_AllKeys(t *testing.T) {
	s := New("/tmp/test-state.json")
	s.Upsert(Entry{PageKey: "key1"})
	s.Upsert(Entry{PageKey: "key2"})
	s.Upsert(Entry{PageKey: "key3"})
	keys := s.AllKeys()
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
	assert.Contains(t, keys, "key3")
}

func TestStore_Save_CreatesParentDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "dir", "state.json")
	s := New(path)
	s.Upsert(Entry{PageKey: "key1", Fingerprint: "abc"})
	require.NoError(t, s.Save())
	_, err := os.Stat(path)
	assert.NoError(t, err)
}
