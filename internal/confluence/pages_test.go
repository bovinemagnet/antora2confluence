package confluence

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPage_ReturnsPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/wiki/api/v2/pages/123", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "body-format=storage")
		json.NewEncoder(w).Encode(Page{ID: "123", Title: "Test Page", SpaceID: "1"})
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	page, err := c.GetPage("123")
	require.NoError(t, err)
	assert.Equal(t, "123", page.ID)
	assert.Equal(t, "Test Page", page.Title)
}

func TestGetChildPages_ReturnsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/wiki/api/v2/pages/100/children", r.URL.Path)
		json.NewEncoder(w).Encode(PageList{Results: []Page{{ID: "101", Title: "Child 1"}, {ID: "102", Title: "Child 2"}}})
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	pages, err := c.GetChildPages("100")
	require.NoError(t, err)
	assert.Len(t, pages, 2)
}

func TestGetPageByTitle_Found(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/wiki/api/v2/pages", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "space-id=1")
		assert.Contains(t, r.URL.RawQuery, "title=Test")
		json.NewEncoder(w).Encode(PageList{Results: []Page{{ID: "42", Title: "Test"}}})
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	page, err := c.GetPageByTitle("1", "Test")
	require.NoError(t, err)
	require.NotNil(t, page)
	assert.Equal(t, "42", page.ID)
}

func TestGetPageByTitle_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(PageList{Results: []Page{}})
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	page, err := c.GetPageByTitle("1", "Missing")
	require.NoError(t, err)
	assert.Nil(t, page)
}

func TestCreatePage_ReturnsCreatedPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/wiki/api/v2/pages", r.URL.Path)
		var req CreatePageRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "New Page", req.Title)
		json.NewEncoder(w).Encode(Page{ID: "200", Title: "New Page", SpaceID: "1"})
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	page, err := c.CreatePage(CreatePageRequest{
		SpaceID: "1", Status: "current", Title: "New Page", ParentID: "100",
		Body: Body{Storage: &Storage{Value: "<p>Hello</p>", Representation: "storage"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "200", page.ID)
}

func TestUpdatePage_ReturnsUpdatedPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/wiki/api/v2/pages/200", r.URL.Path)
		var req UpdatePageRequest
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "Updated Page", req.Title)
		assert.Equal(t, 2, req.Version.Number)
		json.NewEncoder(w).Encode(Page{ID: "200", Title: "Updated Page"})
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	page, err := c.UpdatePage("200", UpdatePageRequest{
		ID: "200", Status: "current", Title: "Updated Page",
		Body: Body{Storage: &Storage{Value: "<p>Updated</p>", Representation: "storage"}},
		Version: Version{Number: 2},
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated Page", page.Title)
}

func TestAddLabels_SendsCorrectPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/wiki/rest/api/content/200/label", r.URL.Path)
		var labels []Label
		json.NewDecoder(r.Body).Decode(&labels)
		assert.Len(t, labels, 2)
		assert.Equal(t, "managed", labels[0].Name)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"results": labels})
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	err := c.AddLabels("200", []string{"managed", "docs"})
	require.NoError(t, err)
}
