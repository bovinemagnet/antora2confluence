package publisher

import (
	"testing"

	"github.com/bovinemagnet/antora2confluence/internal/confluence"
	"github.com/bovinemagnet/antora2confluence/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockPage(id, title, spaceID string) confluence.Page {
	return confluence.Page{ID: id, Title: title, SpaceID: spaceID, Version: &confluence.Version{Number: 1}}
}

func TestEnsureHierarchy_CreatesComponentVersionModule(t *testing.T) {
	mock := newMockClient()
	pub := New(mock, "root-1", "space1", nil)

	source := &model.ContentSource{
		SiteKey: "test",
		Components: []model.Component{
			{
				Name: "billing",
				Versions: []model.Version{
					{
						Name: "1.0",
						Modules: []model.Module{
							{Name: "ROOT"},
							{Name: "api"},
						},
					},
				},
			},
		},
	}

	h, err := pub.EnsureHierarchy(source)
	require.NoError(t, err)

	// Should have created: billing, billing/1.0, billing/1.0/ROOT, billing/1.0/api
	assert.NotEmpty(t, h["billing"])
	assert.NotEmpty(t, h["billing/1.0"])
	assert.NotEmpty(t, h["billing/1.0/ROOT"])
	assert.NotEmpty(t, h["billing/1.0/api"])

	// Total pages created: 4 (component + version + 2 modules)
	assert.Len(t, mock.pages, 4)
}

func TestEnsureHierarchy_ReusesExistingPages(t *testing.T) {
	mock := newMockClient()
	// Pre-create the billing page
	p := mockPage("existing-billing", "billing", "space1")
	mock.pages["existing-billing"] = &p
	pub := New(mock, "root-1", "space1", nil)

	source := &model.ContentSource{
		Components: []model.Component{
			{
				Name: "billing",
				Versions: []model.Version{
					{Name: "1.0", Modules: []model.Module{{Name: "ROOT"}}},
				},
			},
		},
	}

	h, err := pub.EnsureHierarchy(source)
	require.NoError(t, err)

	assert.Equal(t, "existing-billing", h["billing"])
	// Only version + module should be newly created (2 new pages)
	// existing-billing was already there + 2 new = 3 total
	assert.Len(t, mock.pages, 3)
}

func TestHierarchyMap_ParentIDForPage(t *testing.T) {
	h := HierarchyMap{
		"billing/1.0/ROOT": "mod-100",
		"billing/1.0/api":  "mod-200",
	}

	assert.Equal(t, "mod-100", h.ParentIDForPage("billing", "1.0", "ROOT"))
	assert.Equal(t, "mod-200", h.ParentIDForPage("billing", "1.0", "api"))
	assert.Equal(t, "", h.ParentIDForPage("billing", "2.0", "ROOT"))
}
