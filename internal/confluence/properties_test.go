package confluence

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPageProperty_Found(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/wiki/rest/api/content/200/property/antora-page-key", r.URL.Path)
		json.NewEncoder(w).Encode(Property{Key: "antora-page-key", Value: "billing/1.0/ROOT/index.adoc", Version: &PropVersion{Number: 1}})
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	prop, err := c.GetPageProperty("200", "antora-page-key")
	require.NoError(t, err)
	require.NotNil(t, prop)
	assert.Equal(t, "antora-page-key", prop.Key)
}

func TestGetPageProperty_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"statusCode":404}`))
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	prop, err := c.GetPageProperty("200", "missing-key")
	require.NoError(t, err)
	assert.Nil(t, prop)
}

func TestSetPageProperty_Creates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/wiki/rest/api/content/200/property", r.URL.Path)
		var prop Property
		json.NewDecoder(r.Body).Decode(&prop)
		assert.Equal(t, "antora-page-key", prop.Key)
		json.NewEncoder(w).Encode(prop)
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	err := c.SetPageProperty("200", Property{Key: "antora-page-key", Value: "billing/1.0/ROOT/index.adoc"})
	require.NoError(t, err)
}

func TestSetPageProperty_Updates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/wiki/rest/api/content/200/property/antora-page-key", r.URL.Path)
		var prop Property
		json.NewDecoder(r.Body).Decode(&prop)
		assert.Equal(t, 2, prop.Version.Number)
		json.NewEncoder(w).Encode(prop)
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	err := c.SetPageProperty("200", Property{Key: "antora-page-key", Value: "billing/1.0/ROOT/index.adoc", Version: &PropVersion{Number: 2}})
	require.NoError(t, err)
}
