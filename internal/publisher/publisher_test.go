package publisher

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/bovinemagnet/antora2confluence/internal/confluence"
	"github.com/bovinemagnet/antora2confluence/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	pages       map[string]*confluence.Page
	properties  map[string]map[string]*confluence.Property
	labels      map[string][]string
	attachments map[string][]string // pageID -> list of filenames
	nextID      int
}

func newMockClient() *mockClient {
	return &mockClient{
		pages:       make(map[string]*confluence.Page),
		properties:  make(map[string]map[string]*confluence.Property),
		labels:      make(map[string][]string),
		attachments: make(map[string][]string),
		nextID:      1000,
	}
}

func (m *mockClient) GetPageByTitle(spaceID, title string) (*confluence.Page, error) {
	for _, p := range m.pages {
		if p.Title == title && p.SpaceID == spaceID {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockClient) CreatePage(req confluence.CreatePageRequest) (*confluence.Page, error) {
	m.nextID++
	id := fmt.Sprintf("%d", m.nextID)
	page := &confluence.Page{
		ID: id, Title: req.Title, SpaceID: req.SpaceID, ParentID: req.ParentID,
		Version: &confluence.Version{Number: 1},
	}
	m.pages[id] = page
	return page, nil
}

func (m *mockClient) GetPage(id string) (*confluence.Page, error) {
	p, ok := m.pages[id]
	if !ok {
		return nil, fmt.Errorf("page %s not found", id)
	}
	return p, nil
}

func (m *mockClient) UpdatePage(id string, req confluence.UpdatePageRequest) (*confluence.Page, error) {
	p, ok := m.pages[id]
	if !ok {
		return nil, fmt.Errorf("page %s not found", id)
	}
	p.Title = req.Title
	p.Version = &confluence.Version{Number: req.Version.Number}
	return p, nil
}

func (m *mockClient) AddLabels(pageID string, labels []string) error {
	m.labels[pageID] = append(m.labels[pageID], labels...)
	return nil
}

func (m *mockClient) SetPageProperty(pageID string, prop confluence.Property) error {
	if m.properties[pageID] == nil {
		m.properties[pageID] = make(map[string]*confluence.Property)
	}
	m.properties[pageID][prop.Key] = &prop
	return nil
}

func (m *mockClient) GetPageProperty(pageID, key string) (*confluence.Property, error) {
	if props, ok := m.properties[pageID]; ok {
		if p, ok := props[key]; ok {
			return p, nil
		}
	}
	return nil, nil
}

func (m *mockClient) UploadAttachment(pageID, filename string, reader io.Reader) error {
	m.attachments[pageID] = append(m.attachments[pageID], filename)
	return nil
}

func TestPublish_CreatesPagesForPlanItems(t *testing.T) {
	mock := newMockClient()
	pub := New(mock, "1", "space1", []string{"managed"})

	plan := model.PublishPlan{
		Items: []model.PlanItem{
			{
				Page:        model.Page{PageKey: "site/billing/1.0/ROOT/index.adoc", Title: "Overview"},
				Action:      model.ActionCreate,
				ParentID:    "100",
				Fingerprint: "abc123",
			},
		},
	}
	rendered := map[string]*model.RenderedPage{
		"site/billing/1.0/ROOT/index.adoc": {Title: "Overview", Body: "<p>Hello</p>"},
	}

	result := pub.Execute(plan, rendered)
	assert.Equal(t, 1, result.Created)
	assert.Equal(t, 0, result.Failed)
}

func TestPublish_UpdatesExistingPages(t *testing.T) {
	mock := newMockClient()
	mock.pages["200"] = &confluence.Page{
		ID: "200", Title: "Overview", SpaceID: "space1",
		Version: &confluence.Version{Number: 1},
	}
	pub := New(mock, "1", "space1", []string{"managed"})

	plan := model.PublishPlan{
		Items: []model.PlanItem{
			{
				Page:         model.Page{PageKey: "site/billing/1.0/ROOT/index.adoc", Title: "Overview"},
				Action:       model.ActionUpdate,
				ConfluenceID: "200",
				Fingerprint:  "def456",
			},
		},
	}
	rendered := map[string]*model.RenderedPage{
		"site/billing/1.0/ROOT/index.adoc": {Title: "Overview", Body: "<p>Updated</p>"},
	}

	result := pub.Execute(plan, rendered)
	assert.Equal(t, 1, result.Updated)
	assert.Equal(t, 0, result.Failed)
}

func TestPublish_SkipsSkippedItems(t *testing.T) {
	mock := newMockClient()
	pub := New(mock, "1", "space1", nil)

	plan := model.PublishPlan{
		Items: []model.PlanItem{
			{Page: model.Page{PageKey: "site/billing/1.0/ROOT/index.adoc"}, Action: model.ActionSkip},
		},
	}

	result := pub.Execute(plan, nil)
	assert.Equal(t, 1, result.Skipped)
}

func TestPublish_SetsLabelsAndProperties(t *testing.T) {
	mock := newMockClient()
	pub := New(mock, "1", "space1", []string{"managed", "docs"})

	plan := model.PublishPlan{
		Items: []model.PlanItem{
			{
				Page:        model.Page{PageKey: "site/billing/1.0/ROOT/index.adoc", Title: "Overview"},
				Action:      model.ActionCreate,
				ParentID:    "100",
				Fingerprint: "abc123",
			},
		},
	}
	rendered := map[string]*model.RenderedPage{
		"site/billing/1.0/ROOT/index.adoc": {Title: "Overview", Body: "<p>Hello</p>"},
	}

	pub.Execute(plan, rendered)

	var createdID string
	for id := range mock.pages {
		createdID = id
	}
	require.NotEmpty(t, createdID)
	assert.Contains(t, mock.labels[createdID], "managed")
	assert.Contains(t, mock.labels[createdID], "docs")
	assert.NotNil(t, mock.properties[createdID]["antora-page-key"])
	assert.NotNil(t, mock.properties[createdID]["antora-fingerprint"])
}

func TestPublish_UploadsImages(t *testing.T) {
	mock := newMockClient()
	pub := New(mock, "1", "space1", nil)

	// Create a temp dir with an image in the expected location
	dir := t.TempDir()
	modDir := filepath.Join(dir, "modules", "ROOT")
	pagesDir := filepath.Join(modDir, "pages")
	imagesDir := filepath.Join(modDir, "assets", "images")
	os.MkdirAll(pagesDir, 0755)
	os.MkdirAll(imagesDir, 0755)
	os.WriteFile(filepath.Join(imagesDir, "logo.png"), []byte("fake-png"), 0644)

	pageAbsPath := filepath.Join(pagesDir, "index.adoc")

	plan := model.PublishPlan{
		Items: []model.PlanItem{
			{
				Page: model.Page{
					PageKey: "site/comp/1.0/ROOT/index.adoc",
					Title:   "Overview",
					AbsPath: pageAbsPath,
				},
				Action:      model.ActionCreate,
				ParentID:    "100",
				Fingerprint: "abc123",
			},
		},
	}
	rendered := map[string]*model.RenderedPage{
		"site/comp/1.0/ROOT/index.adoc": {
			SourcePage: model.Page{AbsPath: pageAbsPath},
			Title:      "Overview",
			Body:       "<p>Hello</p>",
			Images:     []string{"logo.png"},
		},
	}

	result := pub.Execute(plan, rendered)
	assert.Equal(t, 1, result.Created)

	// Find the created page ID
	var createdID string
	for id := range mock.pages {
		createdID = id
	}
	require.NotEmpty(t, createdID)
	assert.Contains(t, mock.attachments[createdID], "logo.png")
}
