package confluence

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadAttachment_SendsMultipart(t *testing.T) {
	var gotContentType string
	var gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/wiki/rest/api/content/200/child/attachment", r.URL.Path)
		assert.Equal(t, "nocheck", r.Header.Get("X-Atlassian-Token"))
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"results":[]}`))
	}))
	defer server.Close()
	c := NewClient(server.URL+"/wiki", "user", "token")
	err := c.UploadAttachment("200", "logo.png", strings.NewReader("fake-png-data"))
	require.NoError(t, err)
	assert.Contains(t, gotContentType, "multipart/form-data")
	assert.Contains(t, gotBody, "logo.png")
	assert.Contains(t, gotBody, "fake-png-data")
}
