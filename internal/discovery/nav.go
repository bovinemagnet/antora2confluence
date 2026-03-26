package discovery

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/bovinemagnet/antora2confluence/internal/model"
)

var navXrefPattern = regexp.MustCompile(`xref:([^\[]+)\[([^\]]*)\]`)

func ParseNav(path string) ([]model.NavEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Parse all lines into flat entries with their levels.
	type rawEntry struct {
		level int
		entry model.NavEntry
	}

	var raw []rawEntry
	order := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		level := 0
		for _, ch := range trimmed {
			if ch == '*' {
				level++
			} else {
				break
			}
		}
		if level == 0 {
			continue
		}

		text := strings.TrimSpace(trimmed[level:])

		entry := model.NavEntry{Order: order}
		order++

		if m := navXrefPattern.FindStringSubmatch(text); m != nil {
			entry.PageRef = m[1]
			entry.Title = m[2]
		} else {
			entry.Title = text
		}

		raw = append(raw, rawEntry{level: level, entry: entry})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(raw) == 0 {
		return nil, nil
	}

	// Build tree using a stack-based approach.
	// The stack holds pointers to the []NavEntry slices at each level.
	var result []model.NavEntry

	// Stack of (level, pointer to parent's Children slice).
	// Level 0 is a virtual root whose children list is &result.
	type frame struct {
		level    int
		children *[]model.NavEntry
	}

	stack := []frame{{level: 0, children: &result}}

	for _, r := range raw {
		// Pop stack until we find a frame with level < r.level.
		for len(stack) > 1 && stack[len(stack)-1].level >= r.level {
			stack = stack[:len(stack)-1]
		}

		parent := stack[len(stack)-1].children
		*parent = append(*parent, r.entry)

		// Push a frame for this entry's children.
		lastIdx := len(*parent) - 1
		stack = append(stack, frame{
			level:    r.level,
			children: &(*parent)[lastIdx].Children,
		})
	}

	return result, nil
}
