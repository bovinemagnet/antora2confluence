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
	Page        Page
	Action      Action
	Reason      string
	ParentID    string
	Fingerprint string
}

type PublishResult struct {
	Created   int
	Updated   int
	Skipped   int
	Failed    int
	Orphaned  int
	Errors    []error
	StartedAt time.Time
	EndedAt   time.Time
}
