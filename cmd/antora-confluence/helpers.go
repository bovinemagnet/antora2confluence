package main

import (
	"github.com/bovinemagnet/antora2confluence/internal/config"
	"github.com/bovinemagnet/antora2confluence/internal/renderer"
)

func buildBackendConfig(cfg *config.Config, sourceRoot string) renderer.BackendConfig {
	bc := renderer.BackendConfig{
		Backend:     cfg.Render.Backend,
		Extensions:  cfg.Render.Extensions,
		DockerImage: cfg.Render.DockerImage,
		SourceRoot:  sourceRoot,
	}

	// CLI flags override config
	if useDocker {
		bc.Backend = "docker"
	}
	if usePodman {
		bc.Backend = "podman"
	}

	return bc
}
