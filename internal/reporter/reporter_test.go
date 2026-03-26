package reporter

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bovinemagnet/antora2confluence/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintPlan_ShowsActions(t *testing.T) {
	var buf bytes.Buffer
	plan := model.PublishPlan{
		Items: []model.PlanItem{
			{Page: model.Page{PageKey: "site/comp/1.0/ROOT/index.adoc", Title: "Overview"}, Action: model.ActionCreate, Reason: "new page"},
			{Page: model.Page{PageKey: "site/comp/1.0/ROOT/about.adoc", Title: "About"}, Action: model.ActionUpdate, Reason: "changed"},
			{Page: model.Page{PageKey: "site/comp/1.0/api/auth.adoc", Title: "Auth"}, Action: model.ActionSkip, Reason: "unchanged"},
			{Page: model.Page{PageKey: "site/comp/1.0/api/old.adoc", Title: "Old"}, Action: model.ActionOrphan, Reason: "deleted"},
		},
	}
	PrintPlan(&buf, plan)
	output := buf.String()
	assert.Contains(t, output, "CREATE")
	assert.Contains(t, output, "UPDATE")
	assert.Contains(t, output, "SKIP")
	assert.Contains(t, output, "ORPHAN")
	assert.Contains(t, output, "Overview")
	assert.Contains(t, output, "1 create")
	assert.Contains(t, output, "1 update")
	assert.Contains(t, output, "1 skip")
	assert.Contains(t, output, "1 orphan")
}

func TestPrintResult_ShowsSummary(t *testing.T) {
	var buf bytes.Buffer
	result := model.PublishResult{
		Created: 3, Updated: 2, Skipped: 10, Failed: 1, Orphaned: 1,
		StartedAt: time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
		EndedAt:   time.Date(2026, 3, 27, 10, 0, 5, 0, time.UTC),
	}
	PrintResult(&buf, result)
	output := buf.String()
	assert.Contains(t, output, "3")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "10")
	assert.Contains(t, output, "1")
}

func TestWriteJSON_WritesValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")
	result := model.PublishResult{
		Created: 2, Updated: 1, Skipped: 5,
		StartedAt: time.Date(2026, 3, 27, 10, 0, 0, 0, time.UTC),
		EndedAt:   time.Date(2026, 3, 27, 10, 0, 3, 0, time.UTC),
	}
	require.NoError(t, WriteJSON(path, result))
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &parsed))
	assert.Equal(t, float64(2), parsed["created"])
	assert.Equal(t, float64(1), parsed["updated"])
	assert.Equal(t, float64(5), parsed["skipped"])
}

func TestPrintPlan_EmptyPlan(t *testing.T) {
	var buf bytes.Buffer
	PrintPlan(&buf, model.PublishPlan{})
	output := buf.String()
	assert.Contains(t, output, "0 create")
}
