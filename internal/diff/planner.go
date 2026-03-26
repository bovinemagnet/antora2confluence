package diff

import (
	"github.com/bovinemagnet/antora2confluence/internal/depgraph"
	"github.com/bovinemagnet/antora2confluence/internal/model"
	"github.com/bovinemagnet/antora2confluence/internal/state"
)

// Plan compares rendered pages against stored state and produces
// a PublishPlan with create/update/skip/orphan actions.
//
// If fullMode is true, all existing pages are marked for update
// regardless of fingerprint changes.
//
// If graph is non-nil, pages whose dependencies changed (e.g. shared
// includes) are also marked for update even if their own fingerprint
// is unchanged.
func Plan(rendered []model.RenderedPage, store *state.Store, fullMode bool, graph *depgraph.Graph) model.PublishPlan {
	var items []model.PlanItem
	seen := make(map[string]bool)

	// First pass: compare fingerprints
	for _, rp := range rendered {
		pageKey := rp.SourcePage.PageKey
		seen[pageKey] = true
		existing, found := store.Lookup(pageKey)

		if !found {
			items = append(items, model.PlanItem{
				Page:        rp.SourcePage,
				Action:      model.ActionCreate,
				Reason:      "new page not in state",
				Fingerprint: rp.Fingerprint,
			})
			continue
		}
		if fullMode {
			items = append(items, model.PlanItem{
				Page:         rp.SourcePage,
				Action:       model.ActionUpdate,
				Reason:       "full mode republish",
				ConfluenceID: existing.ConfluenceID,
				ParentID:     existing.ParentID,
				Fingerprint:  rp.Fingerprint,
			})
			continue
		}
		if existing.Fingerprint == rp.Fingerprint {
			items = append(items, model.PlanItem{
				Page:         rp.SourcePage,
				Action:       model.ActionSkip,
				Reason:       "unchanged fingerprint",
				ConfluenceID: existing.ConfluenceID,
				ParentID:     existing.ParentID,
				Fingerprint:  rp.Fingerprint,
			})
			continue
		}
		items = append(items, model.PlanItem{
			Page:         rp.SourcePage,
			Action:       model.ActionUpdate,
			Reason:       "fingerprint changed",
			ConfluenceID: existing.ConfluenceID,
			ParentID:     existing.ParentID,
			Fingerprint:  rp.Fingerprint,
		})
	}

	// Second pass: propagate dependency changes
	if graph != nil {
		// Collect all dependencies of changed/created pages
		changedDeps := make(map[string]bool)
		for _, item := range items {
			if item.Action == model.ActionCreate || item.Action == model.ActionUpdate {
				for _, dep := range graph.DependenciesOf(item.Page.PageKey) {
					changedDeps[dep] = true
				}
			}
		}

		// Upgrade skipped pages that share a dependency with a changed page
		if len(changedDeps) > 0 {
			for i := range items {
				if items[i].Action != model.ActionSkip {
					continue
				}
				for _, dep := range graph.DependenciesOf(items[i].Page.PageKey) {
					if changedDeps[dep] {
						items[i].Action = model.ActionUpdate
						items[i].Reason = "dependency changed"
						break
					}
				}
			}
		}
	}

	// Detect orphans
	for _, key := range store.AllKeys() {
		if !seen[key] {
			entry, _ := store.Lookup(key)
			items = append(items, model.PlanItem{
				Page:         model.Page{PageKey: key, Title: entry.Title},
				Action:       model.ActionOrphan,
				Reason:       "source page no longer exists",
				ConfluenceID: entry.ConfluenceID,
			})
		}
	}

	return model.PublishPlan{Items: items}
}
