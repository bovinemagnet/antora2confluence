package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/bovinemagnet/antora2confluence/internal/model"
	"gopkg.in/yaml.v3"
)

type antoraDescriptor struct {
	Name    string   `yaml:"name"`
	Version string   `yaml:"version"`
	Nav     []string `yaml:"nav"`
}

func Scan(root string, siteKey string) (*model.ContentSource, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolving path %s: %w", root, err)
	}

	desc, err := readAntoraYml(absRoot)
	if err != nil {
		return nil, err
	}

	modules, err := discoverModules(absRoot, siteKey, desc.Name, desc.Version)
	if err != nil {
		return nil, fmt.Errorf("discovering modules: %w", err)
	}

	source := &model.ContentSource{
		Root:    absRoot,
		SiteKey: siteKey,
		Components: []model.Component{
			{
				Name: desc.Name,
				Versions: []model.Version{
					{
						Name:    desc.Version,
						Modules: modules,
					},
				},
			},
		},
	}

	return source, nil
}

func readAntoraYml(root string) (*antoraDescriptor, error) {
	data, err := os.ReadFile(filepath.Join(root, "antora.yml"))
	if err != nil {
		return nil, fmt.Errorf("reading antora.yml: %w", err)
	}

	var desc antoraDescriptor
	if err := yaml.Unmarshal(data, &desc); err != nil {
		return nil, fmt.Errorf("parsing antora.yml: %w", err)
	}

	if desc.Name == "" {
		return nil, fmt.Errorf("antora.yml: required field 'name' is missing")
	}

	if desc.Version == "" {
		desc.Version = "unversioned"
	}

	return &desc, nil
}

func discoverModules(root, siteKey, compName, version string) ([]model.Module, error) {
	modulesDir := filepath.Join(root, "modules")
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return nil, fmt.Errorf("reading modules directory: %w", err)
	}

	var modules []model.Module
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		modName := entry.Name()
		modPath := filepath.Join(modulesDir, modName)

		pages, err := discoverPages(modPath, siteKey, compName, version, modName)
		if err != nil {
			return nil, fmt.Errorf("discovering pages in module %s: %w", modName, err)
		}

		nav, err := ParseNav(filepath.Join(modPath, "nav.adoc"))
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("parsing nav for module %s: %w", modName, err)
		}

		modules = append(modules, model.Module{
			Name:  modName,
			Pages: pages,
			Nav:   nav,
		})
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name < modules[j].Name
	})

	return modules, nil
}

func discoverPages(modPath, siteKey, compName, version, modName string) ([]model.Page, error) {
	pagesDir := filepath.Join(modPath, "pages")
	if _, err := os.Stat(pagesDir); os.IsNotExist(err) {
		return nil, nil
	}

	var pages []model.Page
	err := filepath.Walk(pagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".adoc" {
			return nil
		}

		relPath, _ := filepath.Rel(pagesDir, path)
		pageKey := fmt.Sprintf("%s/%s/%s/%s/%s", siteKey, compName, version, modName, relPath)

		page, err := ParsePage(path, relPath, pageKey)
		if err != nil {
			return fmt.Errorf("parsing page %s: %w", relPath, err)
		}

		pages = append(pages, *page)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].RelPath < pages[j].RelPath
	})

	return pages, nil
}
