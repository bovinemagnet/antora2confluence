package model

type ContentSource struct {
	Root       string
	SiteKey    string
	Components []Component
}

func (cs *ContentSource) AllPages() []Page {
	var pages []Page
	for _, comp := range cs.Components {
		for _, ver := range comp.Versions {
			for _, mod := range ver.Modules {
				pages = append(pages, mod.Pages...)
			}
		}
	}
	return pages
}

type Component struct {
	Name     string
	Versions []Version
}

type Version struct {
	Name    string
	Modules []Module
}

type Module struct {
	Name  string
	Pages []Page
	Nav   []NavEntry
}

type Page struct {
	RelPath     string
	AbsPath     string
	Title       string
	PageKey     string
	ExplicitKey string
	Includes    []string
	Images      []string
	XRefs       []string
}

func (p *Page) EffectiveKey() string {
	if p.ExplicitKey != "" {
		return p.ExplicitKey
	}
	return p.PageKey
}

type NavEntry struct {
	Title    string
	PageRef  string
	Children []NavEntry
	Order    int
}
