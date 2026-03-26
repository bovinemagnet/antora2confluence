package renderer

import (
	"bytes"
	"fmt"
	"os/exec"
)

// ConvertToHTML converts an AsciiDoc file to HTML5 by shelling out
// to the asciidoctor CLI. Returns the HTML body content (no standalone document wrapper).
func ConvertToHTML(absPath string) (string, error) {
	cmd := exec.Command(
		"asciidoctor",
		"-b", "html5",
		"-s",
		"-o", "-",
		absPath,
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("asciidoctor failed for %s: %w\nstderr: %s", absPath, err, stderr.String())
	}

	return stdout.String(), nil
}
