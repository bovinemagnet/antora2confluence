package diff

import (
	"testing"

	"github.com/bovinemagnet/antora2confluence/internal/model"
	"github.com/bovinemagnet/antora2confluence/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlan_NewPage_ReturnsCreate(t *testing.T) {
	store := state.New("/tmp/test.json")
	rendered := []model.RenderedPage{{SourcePage: model.Page{PageKey: "site/comp/1.0/ROOT/index.adoc", Title: "Overview"}, Fingerprint: "abc123"}}
	plan := Plan(rendered, store, false)
	require.Len(t, plan.Items, 1)
	assert.Equal(t, model.ActionCreate, plan.Items[0].Action)
	assert.Equal(t, "abc123", plan.Items[0].Fingerprint)
	assert.Contains(t, plan.Items[0].Reason, "new")
}

func TestPlan_UnchangedPage_ReturnsSkip(t *testing.T) {
	store := state.New("/tmp/test.json")
	store.Upsert(state.Entry{PageKey: "site/comp/1.0/ROOT/index.adoc", ConfluenceID: "100", Fingerprint: "abc123"})
	rendered := []model.RenderedPage{{SourcePage: model.Page{PageKey: "site/comp/1.0/ROOT/index.adoc"}, Fingerprint: "abc123"}}
	plan := Plan(rendered, store, false)
	require.Len(t, plan.Items, 1)
	assert.Equal(t, model.ActionSkip, plan.Items[0].Action)
	assert.Contains(t, plan.Items[0].Reason, "unchanged")
}

func TestPlan_ChangedPage_ReturnsUpdate(t *testing.T) {
	store := state.New("/tmp/test.json")
	store.Upsert(state.Entry{PageKey: "site/comp/1.0/ROOT/index.adoc", ConfluenceID: "100", Fingerprint: "old-fingerprint"})
	rendered := []model.RenderedPage{{SourcePage: model.Page{PageKey: "site/comp/1.0/ROOT/index.adoc"}, Fingerprint: "new-fingerprint"}}
	plan := Plan(rendered, store, false)
	require.Len(t, plan.Items, 1)
	assert.Equal(t, model.ActionUpdate, plan.Items[0].Action)
	assert.Equal(t, "100", plan.Items[0].ConfluenceID)
	assert.Contains(t, plan.Items[0].Reason, "changed")
}

func TestPlan_OrphanedPage_ReturnsOrphan(t *testing.T) {
	store := state.New("/tmp/test.json")
	store.Upsert(state.Entry{PageKey: "site/comp/1.0/ROOT/deleted.adoc", ConfluenceID: "999", Fingerprint: "old"})
	rendered := []model.RenderedPage{{SourcePage: model.Page{PageKey: "site/comp/1.0/ROOT/index.adoc"}, Fingerprint: "abc"}}
	plan := Plan(rendered, store, false)
	var orphan, create *model.PlanItem
	for i := range plan.Items {
		switch plan.Items[i].Action {
		case model.ActionOrphan:
			orphan = &plan.Items[i]
		case model.ActionCreate:
			create = &plan.Items[i]
		}
	}
	require.NotNil(t, orphan)
	assert.Equal(t, "site/comp/1.0/ROOT/deleted.adoc", orphan.Page.PageKey)
	require.NotNil(t, create)
}

func TestPlan_FullMode_ForcesUpdateOnUnchanged(t *testing.T) {
	store := state.New("/tmp/test.json")
	store.Upsert(state.Entry{PageKey: "site/comp/1.0/ROOT/index.adoc", ConfluenceID: "100", Fingerprint: "abc123"})
	rendered := []model.RenderedPage{{SourcePage: model.Page{PageKey: "site/comp/1.0/ROOT/index.adoc"}, Fingerprint: "abc123"}}
	plan := Plan(rendered, store, true)
	require.Len(t, plan.Items, 1)
	assert.Equal(t, model.ActionUpdate, plan.Items[0].Action)
	assert.Contains(t, plan.Items[0].Reason, "full")
}

func TestPlan_MixedActions(t *testing.T) {
	store := state.New("/tmp/test.json")
	store.Upsert(state.Entry{PageKey: "existing-unchanged", ConfluenceID: "1", Fingerprint: "fp1"})
	store.Upsert(state.Entry{PageKey: "existing-changed", ConfluenceID: "2", Fingerprint: "old-fp"})
	store.Upsert(state.Entry{PageKey: "orphaned", ConfluenceID: "3", Fingerprint: "fp3"})
	rendered := []model.RenderedPage{
		{SourcePage: model.Page{PageKey: "existing-unchanged"}, Fingerprint: "fp1"},
		{SourcePage: model.Page{PageKey: "existing-changed"}, Fingerprint: "new-fp"},
		{SourcePage: model.Page{PageKey: "brand-new"}, Fingerprint: "fp-new"},
	}
	plan := Plan(rendered, store, false)
	actions := make(map[model.Action]int)
	for _, item := range plan.Items {
		actions[item.Action]++
	}
	assert.Equal(t, 1, actions[model.ActionSkip])
	assert.Equal(t, 1, actions[model.ActionUpdate])
	assert.Equal(t, 1, actions[model.ActionCreate])
	assert.Equal(t, 1, actions[model.ActionOrphan])
}
