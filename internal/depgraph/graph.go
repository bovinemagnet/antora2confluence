package depgraph

import "github.com/bovinemagnet/antora2confluence/internal/model"

type Graph struct {
	depToPages map[string]map[string]bool
	pageToDeps map[string]map[string]bool
}

func Build(pages []model.RenderedPage) *Graph {
	g := &Graph{
		depToPages: make(map[string]map[string]bool),
		pageToDeps: make(map[string]map[string]bool),
	}
	for _, rp := range pages {
		pageKey := rp.SourcePage.PageKey
		for _, inc := range rp.Includes {
			g.addEdge(pageKey, inc)
		}
		for _, img := range rp.Images {
			g.addEdge(pageKey, img)
		}
	}
	return g
}

func (g *Graph) addEdge(pageKey, dep string) {
	if g.depToPages[dep] == nil {
		g.depToPages[dep] = make(map[string]bool)
	}
	g.depToPages[dep][pageKey] = true
	if g.pageToDeps[pageKey] == nil {
		g.pageToDeps[pageKey] = make(map[string]bool)
	}
	g.pageToDeps[pageKey][dep] = true
}

func (g *Graph) AffectedBy(dep string) []string {
	pages := g.depToPages[dep]
	result := make([]string, 0, len(pages))
	for k := range pages {
		result = append(result, k)
	}
	return result
}

func (g *Graph) DependenciesOf(pageKey string) []string {
	deps := g.pageToDeps[pageKey]
	result := make([]string, 0, len(deps))
	for k := range deps {
		result = append(result, k)
	}
	return result
}
