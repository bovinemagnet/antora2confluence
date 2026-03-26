package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Entry struct {
	PageKey      string    `json:"pageKey"`
	ConfluenceID string    `json:"confluenceId"`
	Title        string    `json:"title"`
	Fingerprint  string    `json:"fingerprint"`
	ParentID     string    `json:"parentId"`
	Version      int       `json:"version"`
	PublishedAt  time.Time `json:"publishedAt"`
}

type Store struct {
	path    string
	entries map[string]Entry
}

func New(path string) *Store {
	return &Store{path: path, entries: make(map[string]Entry)}
}

func (s *Store) Load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading state file: %w", err)
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("parsing state file: %w", err)
	}
	s.entries = make(map[string]Entry, len(entries))
	for _, e := range entries {
		s.entries[e.PageKey] = e
	}
	return nil
}

func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}
	entries := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling state: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("writing state file: %w", err)
	}
	return nil
}

func (s *Store) Lookup(pageKey string) (Entry, bool) {
	e, ok := s.entries[pageKey]
	return e, ok
}

func (s *Store) Upsert(entry Entry) {
	s.entries[entry.PageKey] = entry
}

func (s *Store) Entries() map[string]Entry {
	return s.entries
}

func (s *Store) AllKeys() []string {
	keys := make([]string, 0, len(s.entries))
	for k := range s.entries {
		keys = append(keys, k)
	}
	return keys
}
