package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPage_EffectiveKey_UsesExplicitKeyWhenSet(t *testing.T) {
	p := Page{
		PageKey:     "site/comp/1.0/ROOT/index.adoc",
		ExplicitKey: "custom-key",
	}
	assert.Equal(t, "custom-key", p.EffectiveKey())
}

func TestPage_EffectiveKey_FallsBackToPageKey(t *testing.T) {
	p := Page{
		PageKey: "site/comp/1.0/ROOT/index.adoc",
	}
	assert.Equal(t, "site/comp/1.0/ROOT/index.adoc", p.EffectiveKey())
}

func TestContentSource_AllPages_ReturnsAllPagesFlattened(t *testing.T) {
	cs := ContentSource{
		SiteKey: "site",
		Components: []Component{
			{
				Name: "billing",
				Versions: []Version{
					{
						Name: "1.0",
						Modules: []Module{
							{
								Name: "ROOT",
								Pages: []Page{
									{PageKey: "site/billing/1.0/ROOT/index.adoc"},
								},
							},
							{
								Name: "api",
								Pages: []Page{
									{PageKey: "site/billing/1.0/api/auth.adoc"},
									{PageKey: "site/billing/1.0/api/errors.adoc"},
								},
							},
						},
					},
				},
			},
		},
	}
	pages := cs.AllPages()
	assert.Len(t, pages, 3)
	assert.Equal(t, "site/billing/1.0/ROOT/index.adoc", pages[0].PageKey)
	assert.Equal(t, "site/billing/1.0/api/auth.adoc", pages[1].PageKey)
	assert.Equal(t, "site/billing/1.0/api/errors.adoc", pages[2].PageKey)
}
