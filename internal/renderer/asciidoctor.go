package renderer

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
)

// BackendConfig configures how asciidoctor is invoked.
type BackendConfig struct {
	Backend     string   // "local", "docker", or "podman"
	Extensions  []string // asciidoctor extensions to require (-r flags)
	DockerImage string   // container image name for docker/podman backends
	SourceRoot  string   // root directory to bind-mount for container backends
}

// DefaultBackendConfig returns a config for local execution with no extensions.
func DefaultBackendConfig() BackendConfig {
	return BackendConfig{Backend: "local"}
}

// ConvertToHTML converts an AsciiDoc file to HTML5 using the configured backend.
func ConvertToHTML(absPath string, cfg BackendConfig) (string, error) {
	args := buildAsciidoctorArgs(absPath, cfg)

	var cmdName string
	var cmdArgs []string

	switch cfg.Backend {
	case "docker", "podman":
		cmdName = cfg.Backend
		sourceRoot := cfg.SourceRoot
		if sourceRoot == "" {
			sourceRoot = filepath.Dir(absPath)
		}
		cmdArgs = append([]string{
			"run", "--rm",
			"-v", sourceRoot + ":" + sourceRoot + ":ro",
			cfg.DockerImage,
		}, args...)
	default: // "local"
		cmdName = "asciidoctor"
		cmdArgs = args
	}

	cmd := exec.Command(cmdName, cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%s failed for %s: %w\nstderr: %s", cmdName, absPath, err, stderr.String())
	}

	return stdout.String(), nil
}

func buildAsciidoctorArgs(absPath string, cfg BackendConfig) []string {
	args := []string{"-b", "html5", "-s"}
	for _, ext := range cfg.Extensions {
		args = append(args, "-r", ext)
	}
	args = append(args, "-o", "-", absPath)
	return args
}
