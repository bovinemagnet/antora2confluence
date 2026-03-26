package diff

import (
	"github.com/bovinemagnet/antora2confluence/internal/model"
	"github.com/bovinemagnet/antora2confluence/internal/state"
)

func Plan(rendered []model.RenderedPage, store *state.Store, fullMode bool) model.PublishPlan {
	var items []model.PlanItem
	seen := make(map[string]bool)

	for _, rp := range rendered {
		pageKey := rp.SourcePage.PageKey
		seen[pageKey] = true
		existing, found := store.Lookup(pageKey)

		if !found {
			items = append(items, model.PlanItem{
				Page: rp.SourcePage, Action: model.ActionCreate,
				Reason: "new page not in state", Fingerprint: rp.Fingerprint,
			})
			continue
		}
		if fullMode {
			items = append(items, model.PlanItem{
				Page: rp.SourcePage, Action: model.ActionUpdate,
				Reason: "full mode republish", ParentID: existing.ConfluenceID, Fingerprint: rp.Fingerprint,
			})
			continue
		}
		if existing.Fingerprint == rp.Fingerprint {
			items = append(items, model.PlanItem{
				Page: rp.SourcePage, Action: model.ActionSkip,
				Reason: "unchanged fingerprint", ParentID: existing.ConfluenceID, Fingerprint: rp.Fingerprint,
			})
			continue
		}
		items = append(items, model.PlanItem{
			Page: rp.SourcePage, Action: model.ActionUpdate,
			Reason: "fingerprint changed", ParentID: existing.ConfluenceID, Fingerprint: rp.Fingerprint,
		})
	}

	for _, key := range store.AllKeys() {
		if !seen[key] {
			entry, _ := store.Lookup(key)
			items = append(items, model.PlanItem{
				Page:     model.Page{PageKey: key, Title: entry.Title},
				Action:   model.ActionOrphan,
				Reason:   "source page no longer exists",
				ParentID: entry.ConfluenceID,
			})
		}
	}

	return model.PublishPlan{Items: items}
}
