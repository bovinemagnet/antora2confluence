package publisher

import (
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/bovinemagnet/antora2confluence/internal/confluence"
	"github.com/bovinemagnet/antora2confluence/internal/model"
)

// ConfluenceAPI defines the subset of the Confluence client used by the publisher.
type ConfluenceAPI interface {
	GetPageByTitle(spaceID, title string) (*confluence.Page, error)
	CreatePage(req confluence.CreatePageRequest) (*confluence.Page, error)
	GetPage(id string) (*confluence.Page, error)
	UpdatePage(id string, req confluence.UpdatePageRequest) (*confluence.Page, error)
	AddLabels(pageID string, labels []string) error
	SetPageProperty(pageID string, prop confluence.Property) error
	GetPageProperty(pageID, key string) (*confluence.Property, error)
	UploadAttachment(pageID, filename string, reader io.Reader) error
}

// Publisher orchestrates creating and updating Confluence pages
// based on a publish plan.
type Publisher struct {
	client  ConfluenceAPI
	spaceID string
	rootID  string
	labels  []string
}

// New creates a new Publisher.
func New(client ConfluenceAPI, rootPageID, spaceID string, labels []string) *Publisher {
	return &Publisher{
		client:  client,
		rootID:  rootPageID,
		spaceID: spaceID,
		labels:  labels,
	}
}

// Execute runs the publish plan, creating/updating pages as needed.
func (p *Publisher) Execute(plan model.PublishPlan, rendered map[string]*model.RenderedPage) model.PublishResult {
	result := model.PublishResult{
		StartedAt: time.Now(),
	}

	for _, item := range plan.Items {
		switch item.Action {
		case model.ActionSkip:
			result.Skipped++
			slog.Debug("Skipping page", "key", item.Page.PageKey)

		case model.ActionCreate:
			if err := p.createPage(item, rendered); err != nil {
				slog.Error("Failed to create page", "key", item.Page.PageKey, "error", err)
				result.Failed++
				result.Errors = append(result.Errors, err)
			} else {
				result.Created++
			}

		case model.ActionUpdate:
			if err := p.updatePage(item, rendered); err != nil {
				slog.Error("Failed to update page", "key", item.Page.PageKey, "error", err)
				result.Failed++
				result.Errors = append(result.Errors, err)
			} else {
				result.Updated++
			}

		case model.ActionOrphan:
			result.Orphaned++
			slog.Info("Orphaned page detected", "key", item.Page.PageKey)
		}
	}

	result.EndedAt = time.Now()
	return result
}

func (p *Publisher) createPage(item model.PlanItem, rendered map[string]*model.RenderedPage) error {
	rp := rendered[item.Page.PageKey]
	if rp == nil {
		return fmt.Errorf("no rendered content for %s", item.Page.PageKey)
	}

	parentID := item.ParentID
	if parentID == "" {
		parentID = p.rootID
	}

	page, err := p.client.CreatePage(confluence.CreatePageRequest{
		SpaceID:  p.spaceID,
		Status:   "current",
		Title:    rp.Title,
		ParentID: parentID,
		Body: confluence.Body{
			Storage: &confluence.Storage{
				Value:          rp.Body,
				Representation: "storage",
			},
		},
	})
	if err != nil {
		return fmt.Errorf("creating page %s: %w", item.Page.PageKey, err)
	}

	slog.Info("Created page", "key", item.Page.PageKey, "id", page.ID, "title", rp.Title)
	return p.setMetadata(page.ID, item)
}

func (p *Publisher) updatePage(item model.PlanItem, rendered map[string]*model.RenderedPage) error {
	rp := rendered[item.Page.PageKey]
	if rp == nil {
		return fmt.Errorf("no rendered content for %s", item.Page.PageKey)
	}

	existing, err := p.client.GetPage(item.ParentID)
	if err != nil {
		return fmt.Errorf("getting page for update %s: %w", item.Page.PageKey, err)
	}

	newVersion := 1
	if existing.Version != nil {
		newVersion = existing.Version.Number + 1
	}

	page, err := p.client.UpdatePage(existing.ID, confluence.UpdatePageRequest{
		ID:     existing.ID,
		Status: "current",
		Title:  rp.Title,
		Body: confluence.Body{
			Storage: &confluence.Storage{
				Value:          rp.Body,
				Representation: "storage",
			},
		},
		Version: confluence.Version{Number: newVersion},
	})
	if err != nil {
		return fmt.Errorf("updating page %s: %w", item.Page.PageKey, err)
	}

	slog.Info("Updated page", "key", item.Page.PageKey, "id", page.ID, "version", newVersion)
	return p.setMetadata(page.ID, item)
}

func (p *Publisher) setMetadata(pageID string, item model.PlanItem) error {
	if len(p.labels) > 0 {
		if err := p.client.AddLabels(pageID, p.labels); err != nil {
			slog.Warn("Failed to set labels", "pageID", pageID, "error", err)
		}
	}

	if err := p.client.SetPageProperty(pageID, confluence.Property{
		Key:   "antora-page-key",
		Value: item.Page.EffectiveKey(),
	}); err != nil {
		slog.Warn("Failed to set page key property", "pageID", pageID, "error", err)
	}

	if err := p.client.SetPageProperty(pageID, confluence.Property{
		Key:   "antora-fingerprint",
		Value: item.Fingerprint,
	}); err != nil {
		slog.Warn("Failed to set fingerprint property", "pageID", pageID, "error", err)
	}

	return nil
}
