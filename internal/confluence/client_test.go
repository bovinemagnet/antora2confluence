package confluence

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_SetsBaseURLAndAuth(t *testing.T) {
	c := NewClient("https://example.atlassian.net/wiki", "user@example.com", "token123")
	assert.Equal(t, "https://example.atlassian.net/wiki", c.baseURL)
	assert.Equal(t, "user@example.com", c.username)
	assert.Equal(t, "token123", c.token)
}

func TestClient_DoRequest_SetsBasicAuth(t *testing.T) {
	var gotUser, gotPass string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, _ = r.BasicAuth()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "myuser", "mytoken")
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	_, err := c.do(req)
	require.NoError(t, err)
	assert.Equal(t, "myuser", gotUser)
	assert.Equal(t, "mytoken", gotPass)
}

func TestClient_DoRequest_RetriesOn429(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count <= 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "user", "token")
	c.maxRetries = 3
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestClient_DoRequest_RetriesOn5xx(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count <= 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "user", "token")
	c.maxRetries = 3
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	resp, err := c.do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestClient_DoRequest_FailsAfterMaxRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c := NewClient(server.URL, "user", "token")
	c.maxRetries = 2
	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	_, err := c.do(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "429")
}

func TestClient_ValidateAuth_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/wiki/api/v2/spaces", r.URL.Path)
		assert.Equal(t, "key=TEST", r.URL.RawQuery)
		json.NewEncoder(w).Encode(SpaceList{Results: []Space{{ID: "1", Key: "TEST"}}})
	}))
	defer server.Close()

	c := NewClient(server.URL+"/wiki", "user", "token")
	spaceID, err := c.ValidateAuth("TEST")
	require.NoError(t, err)
	assert.Equal(t, "1", spaceID)
}

func TestClient_ValidateAuth_InvalidSpace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SpaceList{Results: []Space{}})
	}))
	defer server.Close()

	c := NewClient(server.URL+"/wiki", "user", "token")
	_, err := c.ValidateAuth("NOPE")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "space")
}
