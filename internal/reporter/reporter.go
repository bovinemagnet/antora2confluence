package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bovinemagnet/antora2confluence/internal/model"
)

func PrintPlan(w io.Writer, plan model.PublishPlan) {
	counts := map[model.Action]int{}
	for _, item := range plan.Items {
		counts[item.Action]++
	}

	fmt.Fprintf(w, "\nPublish Plan\n")
	fmt.Fprintf(w, "============\n\n")

	for _, item := range plan.Items {
		action := strings.ToUpper(string(item.Action))
		fmt.Fprintf(w, "  %-8s %s [%s]\n", action, item.Page.Title, item.Page.PageKey)
		if item.Reason != "" {
			fmt.Fprintf(w, "           Reason: %s\n", item.Reason)
		}
	}

	fmt.Fprintf(w, "\nSummary: %d create, %d update, %d skip, %d orphan\n",
		counts[model.ActionCreate], counts[model.ActionUpdate],
		counts[model.ActionSkip], counts[model.ActionOrphan])
}

func PrintResult(w io.Writer, result model.PublishResult) {
	fmt.Fprintf(w, "\nPublish Result\n")
	fmt.Fprintf(w, "==============\n\n")
	fmt.Fprintf(w, "  Created:  %d\n", result.Created)
	fmt.Fprintf(w, "  Updated:  %d\n", result.Updated)
	fmt.Fprintf(w, "  Skipped:  %d\n", result.Skipped)
	fmt.Fprintf(w, "  Failed:   %d\n", result.Failed)
	fmt.Fprintf(w, "  Orphaned: %d\n", result.Orphaned)

	duration := result.EndedAt.Sub(result.StartedAt)
	if duration == 0 {
		fmt.Fprintf(w, "  Duration: 0s\n")
	} else {
		fmt.Fprintf(w, "  Duration: %s\n", duration.Round(duration/100))
	}

	if len(result.Errors) > 0 {
		fmt.Fprintf(w, "\nErrors:\n")
		for _, err := range result.Errors {
			fmt.Fprintf(w, "  - %s\n", err)
		}
	}
}

type jsonReport struct {
	Created   int      `json:"created"`
	Updated   int      `json:"updated"`
	Skipped   int      `json:"skipped"`
	Failed    int      `json:"failed"`
	Orphaned  int      `json:"orphaned"`
	Errors    []string `json:"errors,omitempty"`
	StartedAt string   `json:"startedAt"`
	EndedAt   string   `json:"endedAt"`
	Duration  string   `json:"duration"`
}

func WriteJSON(path string, result model.PublishResult) error {
	errs := make([]string, len(result.Errors))
	for i, e := range result.Errors {
		errs[i] = e.Error()
	}

	report := jsonReport{
		Created: result.Created, Updated: result.Updated, Skipped: result.Skipped,
		Failed: result.Failed, Orphaned: result.Orphaned, Errors: errs,
		StartedAt: result.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		EndedAt:   result.EndedAt.Format("2006-01-02T15:04:05Z07:00"),
		Duration:  result.EndedAt.Sub(result.StartedAt).String(),
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling report: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}
