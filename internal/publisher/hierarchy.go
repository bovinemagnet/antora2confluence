package publisher

import (
	"fmt"
	"log/slog"

	"github.com/bovinemagnet/antora2confluence/internal/confluence"
	"github.com/bovinemagnet/antora2confluence/internal/model"
)

// HierarchyMap maps a hierarchy path (e.g. "billing/1.0/ROOT") to its Confluence page ID.
type HierarchyMap map[string]string

// EnsureHierarchy creates placeholder pages for the component/version/module
// hierarchy and returns a map from module path to Confluence page ID.
func (p *Publisher) EnsureHierarchy(source *model.ContentSource) (HierarchyMap, error) {
	h := make(HierarchyMap)

	for _, comp := range source.Components {
		compID, err := p.ensurePage(p.rootID, comp.Name)
		if err != nil {
			return nil, fmt.Errorf("ensuring component page %q: %w", comp.Name, err)
		}
		h[comp.Name] = compID

		for _, ver := range comp.Versions {
			verID, err := p.ensurePage(compID, ver.Name)
			if err != nil {
				return nil, fmt.Errorf("ensuring version page %q/%q: %w", comp.Name, ver.Name, err)
			}
			verPath := comp.Name + "/" + ver.Name
			h[verPath] = verID

			for _, mod := range ver.Modules {
				modID, err := p.ensurePage(verID, mod.Name)
				if err != nil {
					return nil, fmt.Errorf("ensuring module page %q/%q/%q: %w", comp.Name, ver.Name, mod.Name, err)
				}
				modPath := verPath + "/" + mod.Name
				h[modPath] = modID
			}
		}
	}

	return h, nil
}

// ensurePage finds or creates a page with the given title under the parent.
func (p *Publisher) ensurePage(parentID, title string) (string, error) {
	existing, err := p.client.GetPageByTitle(p.spaceID, title)
	if err != nil {
		return "", err
	}
	if existing != nil {
		slog.Debug("Hierarchy page exists", "title", title, "id", existing.ID)
		return existing.ID, nil
	}

	page, err := p.client.CreatePage(confluence.CreatePageRequest{
		SpaceID:  p.spaceID,
		Status:   "current",
		Title:    title,
		ParentID: parentID,
		Body: confluence.Body{
			Storage: &confluence.Storage{
				Value:          fmt.Sprintf("<p>Auto-generated index page for <strong>%s</strong>.</p>", title),
				Representation: "storage",
			},
		},
	})
	if err != nil {
		return "", err
	}

	slog.Info("Created hierarchy page", "title", title, "id", page.ID, "parent", parentID)
	return page.ID, nil
}

// ParentIDForPage returns the Confluence page ID that should be the parent
// of a given page, based on its component/version/module path.
func (h HierarchyMap) ParentIDForPage(compName, verName, modName string) string {
	modPath := compName + "/" + verName + "/" + modName
	if id, ok := h[modPath]; ok {
		return id
	}
	return ""
}
