package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Confluence ConfluenceConfig `mapstructure:"confluence"`
	Source     SourceConfig     `mapstructure:"source"`
	Publish    PublishConfig    `mapstructure:"publish"`
	Sync       SyncConfig       `mapstructure:"sync"`
	Render     RenderConfig     `mapstructure:"render"`
	DryRun     bool             `mapstructure:"dryRun"`
	Verbose    bool             `mapstructure:"verbose"`
	ReportFile string           `mapstructure:"reportFile"`
}

type ConfluenceConfig struct {
	BaseURL      string     `mapstructure:"baseUrl"`
	SpaceKey     string     `mapstructure:"spaceKey"`
	ParentPageID string     `mapstructure:"parentPageId"`
	Auth         AuthConfig `mapstructure:"auth"`
}

type AuthConfig struct {
	Mode        string `mapstructure:"mode"`
	UsernameEnv string `mapstructure:"usernameEnv"`
	TokenEnv    string `mapstructure:"tokenEnv"`
}

type SourceConfig struct {
	AntoraRoot string `mapstructure:"antoraRoot"`
	SiteKey    string `mapstructure:"siteKey"`
}

type PublishConfig struct {
	Hierarchy        string   `mapstructure:"hierarchy"`
	VersionMode      string   `mapstructure:"versionMode"`
	CreateIndexPages bool     `mapstructure:"createIndexPages"`
	ApplyLabels      []string `mapstructure:"applyLabels"`
}

type SyncConfig struct {
	Mode           string `mapstructure:"mode"`
	StateFile      string `mapstructure:"stateFile"`
	OrphanStrategy string `mapstructure:"orphanStrategy"`
	Strict         bool   `mapstructure:"strict"`
}

type RenderConfig struct {
	Backend              string   `mapstructure:"backend"`
	Extensions           []string `mapstructure:"extensions"`
	MermaidMode          string   `mapstructure:"mermaidMode"`
	DockerImage          string   `mapstructure:"dockerImage"`
	FailOnUnresolvedXref bool     `mapstructure:"failOnUnresolvedXref"`
	UploadImages         bool     `mapstructure:"uploadImages"`
}

func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetDefault("publish.hierarchy", "component-version-module-page")
	v.SetDefault("publish.versionMode", "hierarchy")
	v.SetDefault("sync.mode", "incremental")
	v.SetDefault("sync.stateFile", ".antora-confluence-state.json")
	v.SetDefault("sync.orphanStrategy", "report")
	v.SetDefault("render.backend", "local")
	v.SetDefault("render.mermaidMode", "passthrough")
	v.SetDefault("render.dockerImage", "antora2confluence/asciidoctor")
	v.SetDefault("render.uploadImages", true)
	v.SetDefault("confluence.auth.mode", "pat")
	v.SetDefault("confluence.auth.usernameEnv", "CONFLUENCE_USER")
	v.SetDefault("confluence.auth.tokenEnv", "CONFLUENCE_TOKEN")

	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// Validate checks that all required configuration fields are set.
func (c *Config) Validate() error {
	if c.Confluence.BaseURL == "" {
		return fmt.Errorf("confluence.baseUrl is required")
	}
	if c.Confluence.SpaceKey == "" {
		return fmt.Errorf("confluence.spaceKey is required")
	}
	if c.Confluence.ParentPageID == "" {
		return fmt.Errorf("confluence.parentPageId is required")
	}
	if c.Source.AntoraRoot == "" {
		return fmt.Errorf("source.antoraRoot is required")
	}
	if c.Source.SiteKey == "" {
		return fmt.Errorf("source.siteKey is required")
	}
	return nil
}
