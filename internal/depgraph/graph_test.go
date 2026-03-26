package depgraph

import (
	"testing"

	"github.com/bovinemagnet/antora2confluence/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBuild_TracksIncludes(t *testing.T) {
	pages := []model.RenderedPage{
		{SourcePage: model.Page{PageKey: "site/comp/1.0/ROOT/index.adoc"}, Includes: []string{"partial$shared.adoc"}},
		{SourcePage: model.Page{PageKey: "site/comp/1.0/ROOT/getting-started.adoc"}, Includes: []string{"partial$shared.adoc"}},
		{SourcePage: model.Page{PageKey: "site/comp/1.0/api/auth.adoc"}},
	}
	g := Build(pages)
	affected := g.AffectedBy("partial$shared.adoc")
	assert.Len(t, affected, 2)
	assert.Contains(t, affected, "site/comp/1.0/ROOT/index.adoc")
	assert.Contains(t, affected, "site/comp/1.0/ROOT/getting-started.adoc")
}

func TestBuild_TracksImages(t *testing.T) {
	pages := []model.RenderedPage{
		{SourcePage: model.Page{PageKey: "site/comp/1.0/ROOT/index.adoc"}, Images: []string{"logo.png"}},
	}
	g := Build(pages)
	affected := g.AffectedBy("logo.png")
	assert.Len(t, affected, 1)
	assert.Contains(t, affected, "site/comp/1.0/ROOT/index.adoc")
}

func TestAffectedBy_UnknownDependency_ReturnsEmpty(t *testing.T) {
	g := Build(nil)
	assert.Empty(t, g.AffectedBy("unknown.adoc"))
}

func TestDependenciesOf_ReturnsAllDeps(t *testing.T) {
	pages := []model.RenderedPage{
		{SourcePage: model.Page{PageKey: "site/comp/1.0/ROOT/index.adoc"}, Includes: []string{"partial$a.adoc", "partial$b.adoc"}, Images: []string{"logo.png"}},
	}
	g := Build(pages)
	deps := g.DependenciesOf("site/comp/1.0/ROOT/index.adoc")
	assert.Len(t, deps, 3)
	assert.Contains(t, deps, "partial$a.adoc")
	assert.Contains(t, deps, "partial$b.adoc")
	assert.Contains(t, deps, "logo.png")
}

func TestDependenciesOf_UnknownPage_ReturnsEmpty(t *testing.T) {
	g := Build(nil)
	assert.Empty(t, g.DependenciesOf("unknown"))
}
