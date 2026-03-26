package discovery

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bovinemagnet/antora2confluence/internal/model"
)

var (
	titlePattern       = regexp.MustCompile(`^= (.+)$`)
	explicitKeyPattern = regexp.MustCompile(`^:confluence-page-key:\s*(.+)$`)
	imagePattern       = regexp.MustCompile(`image::?([^\[]+)\[`)
	xrefPattern        = regexp.MustCompile(`xref:([^\[]+)\[`)
	includePattern     = regexp.MustCompile(`^include::([^\[]+)\[`)
)

func ParsePage(absPath, relPath, pageKey string) (*model.Page, error) {
	file, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", absPath, err)
	}
	defer file.Close()

	page := &model.Page{
		RelPath: relPath,
		AbsPath: absPath,
		PageKey: pageKey,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if page.Title == "" {
			if m := titlePattern.FindStringSubmatch(line); m != nil {
				page.Title = m[1]
			}
		}

		if m := explicitKeyPattern.FindStringSubmatch(line); m != nil {
			page.ExplicitKey = strings.TrimSpace(m[1])
		}

		for _, m := range imagePattern.FindAllStringSubmatch(line, -1) {
			page.Images = append(page.Images, m[1])
		}

		for _, m := range xrefPattern.FindAllStringSubmatch(line, -1) {
			page.XRefs = append(page.XRefs, m[1])
		}

		if m := includePattern.FindStringSubmatch(line); m != nil {
			page.Includes = append(page.Includes, m[1])
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading %s: %w", absPath, err)
	}

	if page.Title == "" {
		name := strings.TrimSuffix(relPath, ".adoc")
		parts := strings.Split(name, "/")
		page.Title = parts[len(parts)-1]
	}

	return page, nil
}
