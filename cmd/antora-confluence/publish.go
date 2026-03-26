package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bovinemagnet/antora2confluence/internal/config"
	"github.com/bovinemagnet/antora2confluence/internal/confluence"
	"github.com/bovinemagnet/antora2confluence/internal/diff"
	"github.com/bovinemagnet/antora2confluence/internal/discovery"
	"github.com/bovinemagnet/antora2confluence/internal/model"
	"github.com/bovinemagnet/antora2confluence/internal/publisher"
	"github.com/bovinemagnet/antora2confluence/internal/renderer"
	"github.com/bovinemagnet/antora2confluence/internal/reporter"
	"github.com/bovinemagnet/antora2confluence/internal/state"
	"github.com/spf13/cobra"
)

var fullSync bool

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish Antora documentation to Confluence",
	Long:  "Discovers Antora content, renders pages, detects changes, and publishes to Confluence. Only changed pages are updated by default.",
	RunE:  runPublish,
}

func init() {
	publishCmd.Flags().BoolVar(&fullSync, "full", false, "republish all pages regardless of changes")
	rootCmd.AddCommand(publishCmd)
}

func runPublish(cmd *cobra.Command, args []string) error {
	if cfgFile == "" {
		return fmt.Errorf("--config flag is required")
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	username := os.Getenv(cfg.Confluence.Auth.UsernameEnv)
	token := os.Getenv(cfg.Confluence.Auth.TokenEnv)
	if username == "" || token == "" {
		return fmt.Errorf("credentials not set: ensure %s and %s environment variables are set",
			cfg.Confluence.Auth.UsernameEnv, cfg.Confluence.Auth.TokenEnv)
	}

	slog.Info("Validating Confluence access", "url", cfg.Confluence.BaseURL, "space", cfg.Confluence.SpaceKey)
	client := confluence.NewClient(cfg.Confluence.BaseURL, username, token)
	spaceID, err := client.ValidateAuth(cfg.Confluence.SpaceKey)
	if err != nil {
		return fmt.Errorf("validating Confluence access: %w", err)
	}
	slog.Info("Confluence access validated", "spaceID", spaceID)

	slog.Info("Scanning Antora content", "root", cfg.Source.AntoraRoot)
	source, err := discovery.Scan(cfg.Source.AntoraRoot, cfg.Source.SiteKey)
	if err != nil {
		return fmt.Errorf("scanning Antora source: %w", err)
	}
	pages := source.AllPages()
	slog.Info("Discovered pages", "count", len(pages))

	slog.Info("Rendering pages")
	r := renderer.New()
	rendered, renderErrs := r.RenderAll(pages)
	if len(renderErrs) > 0 {
		slog.Warn("Some pages failed to render", "failures", len(renderErrs))
		if cfg.Sync.Strict {
			return fmt.Errorf("%d pages failed to render in strict mode", len(renderErrs))
		}
	}

	store := state.New(cfg.Sync.StateFile)
	if err := store.Load(); err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	isFullMode := fullSync || cfg.Sync.Mode == "full"
	plan := diff.Plan(rendered, store, isFullMode)

	reporter.PrintPlan(os.Stdout, plan)

	slog.Info("Publishing pages")
	renderedMap := makeRenderedMap(rendered)
	pub := publisher.New(client, cfg.Confluence.ParentPageID, spaceID, cfg.Publish.ApplyLabels)
	result := pub.Execute(plan, renderedMap)

	for _, item := range plan.Items {
		if item.Action == model.ActionCreate || item.Action == model.ActionUpdate {
			store.Upsert(state.Entry{
				PageKey:     item.Page.PageKey,
				Fingerprint: item.Fingerprint,
				Title:       item.Page.Title,
			})
		}
	}
	if err := store.Save(); err != nil {
		slog.Error("Failed to save state", "error", err)
	}

	reporter.PrintResult(os.Stdout, result)
	if cfg.ReportFile != "" {
		if err := reporter.WriteJSON(cfg.ReportFile, result); err != nil {
			slog.Error("Failed to write JSON report", "error", err)
		}
	}

	if result.Failed > 0 {
		return fmt.Errorf("%d pages failed to publish", result.Failed)
	}
	return nil
}

func makeRenderedMap(rendered []model.RenderedPage) map[string]*model.RenderedPage {
	m := make(map[string]*model.RenderedPage, len(rendered))
	for i := range rendered {
		m[rendered[i].SourcePage.PageKey] = &rendered[i]
	}
	return m
}
