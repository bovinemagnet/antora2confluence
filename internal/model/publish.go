package model

import "time"

type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionSkip   Action = "skip"
	ActionOrphan Action = "orphan"
)

type PublishPlan struct {
	Items []PlanItem
}

type PlanItem struct {
	Page         Page
	Action       Action
	Reason       string
	ParentID     string // Confluence parent page ID (for hierarchy)
	ConfluenceID string // Confluence page ID (for updates)
	Fingerprint  string
}

// PublishedPage records a page that was successfully created or updated.
type PublishedPage struct {
	PageKey      string
	ConfluenceID string
	Title        string
	Fingerprint  string
	ParentID     string
	Version      int
}

type PublishResult struct {
	Created        int
	Updated        int
	Skipped        int
	Failed         int
	Orphaned       int
	Errors         []error
	PublishedPages []PublishedPage
	StartedAt      time.Time
	EndedAt        time.Time
}
