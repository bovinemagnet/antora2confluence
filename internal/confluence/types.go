package confluence

import "time"

// Page represents a Confluence page from the v2 API.
type Page struct {
	ID       string   `json:"id"`
	Status   string   `json:"status"`
	Title    string   `json:"title"`
	SpaceID  string   `json:"spaceId"`
	ParentID string   `json:"parentId"`
	Version  *Version `json:"version,omitempty"`
	Body     *Body    `json:"body,omitempty"`
}

type Version struct {
	Number    int       `json:"number"`
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}

type Body struct {
	Storage *Storage `json:"storage,omitempty"`
}

type Storage struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

type CreatePageRequest struct {
	SpaceID  string `json:"spaceId"`
	Status   string `json:"status"`
	Title    string `json:"title"`
	ParentID string `json:"parentId"`
	Body     Body   `json:"body"`
}

type UpdatePageRequest struct {
	ID      string  `json:"id"`
	Status  string  `json:"status"`
	Title   string  `json:"title"`
	Body    Body    `json:"body"`
	Version Version `json:"version"`
}

type PageList struct {
	Results []Page `json:"results"`
	Links   *Links `json:"_links,omitempty"`
}

type Links struct {
	Next string `json:"next,omitempty"`
}

type Label struct {
	Prefix string `json:"prefix"`
	Name   string `json:"name"`
	ID     string `json:"id,omitempty"`
}

type Property struct {
	Key     string       `json:"key"`
	Value   interface{}  `json:"value"`
	Version *PropVersion `json:"version,omitempty"`
}

type PropVersion struct {
	Number    int  `json:"number"`
	MinorEdit bool `json:"minorEdit"`
}

type Space struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type SpaceList struct {
	Results []Space `json:"results"`
}
