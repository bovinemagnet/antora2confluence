package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_ValidYAML_ReturnsConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
confluence:
  baseUrl: https://example.atlassian.net/wiki
  spaceKey: ENG
  parentPageId: "123456"
  auth:
    mode: pat
    usernameEnv: CONFLUENCE_USER
    tokenEnv: CONFLUENCE_TOKEN
source:
  antoraRoot: ./docs
  siteKey: my-site
publish:
  hierarchy: component-version-module-page
  versionMode: hierarchy
  createIndexPages: true
  applyLabels:
    - managed-by-antora-confluence
sync:
  mode: incremental
  stateFile: .state.json
  orphanStrategy: report
  strict: false
render:
  failOnUnresolvedXref: true
  uploadImages: true
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "https://example.atlassian.net/wiki", cfg.Confluence.BaseURL)
	assert.Equal(t, "ENG", cfg.Confluence.SpaceKey)
	assert.Equal(t, "123456", cfg.Confluence.ParentPageID)
	assert.Equal(t, "pat", cfg.Confluence.Auth.Mode)
	assert.Equal(t, "./docs", cfg.Source.AntoraRoot)
	assert.Equal(t, "my-site", cfg.Source.SiteKey)
	assert.Equal(t, "incremental", cfg.Sync.Mode)
	assert.True(t, cfg.Render.UploadImages)
}

func TestLoad_MissingFile_ReturnsError(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	require.Error(t, err)
}

func TestLoad_Defaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "minimal.yaml")
	content := `
confluence:
  baseUrl: https://example.atlassian.net/wiki
  spaceKey: ENG
  parentPageId: "123"
source:
  antoraRoot: ./docs
  siteKey: test
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	require.NoError(t, err)
	assert.Equal(t, "incremental", cfg.Sync.Mode)
	assert.Equal(t, ".antora-confluence-state.json", cfg.Sync.StateFile)
	assert.Equal(t, "report", cfg.Sync.OrphanStrategy)
	assert.Equal(t, "component-version-module-page", cfg.Publish.Hierarchy)
}

func TestValidate_ValidConfig_NoError(t *testing.T) {
	cfg := &Config{
		Confluence: ConfluenceConfig{BaseURL: "https://example.atlassian.net/wiki", SpaceKey: "ENG", ParentPageID: "123"},
		Source:     SourceConfig{AntoraRoot: "./docs", SiteKey: "test"},
	}
	require.NoError(t, cfg.Validate())
}

func TestValidate_MissingBaseURL_ReturnsError(t *testing.T) {
	cfg := &Config{
		Confluence: ConfluenceConfig{SpaceKey: "ENG", ParentPageID: "123"},
		Source:     SourceConfig{AntoraRoot: "./docs", SiteKey: "test"},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "baseUrl")
}

func TestValidate_MissingSpaceKey_ReturnsError(t *testing.T) {
	cfg := &Config{
		Confluence: ConfluenceConfig{BaseURL: "https://x.atlassian.net/wiki", ParentPageID: "123"},
		Source:     SourceConfig{AntoraRoot: "./docs", SiteKey: "test"},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "spaceKey")
}

func TestValidate_MissingSiteKey_ReturnsError(t *testing.T) {
	cfg := &Config{
		Confluence: ConfluenceConfig{BaseURL: "https://x.atlassian.net/wiki", SpaceKey: "ENG", ParentPageID: "123"},
		Source:     SourceConfig{AntoraRoot: "./docs"},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "siteKey")
}
