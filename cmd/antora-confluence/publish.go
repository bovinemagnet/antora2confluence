package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/bovinemagnet/antora2confluence/internal/config"
	"github.com/bovinemagnet/antora2confluence/internal/confluence"
	"github.com/bovinemagnet/antora2confluence/internal/depgraph"
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

	// 4. Build dependency graph
	graph := depgraph.Build(rendered)

	// 5. Load state and compute diff
	store := state.New(cfg.Sync.StateFile)
	if err := store.Load(); err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	isFullMode := fullSync || cfg.Sync.Mode == "full"
	plan := diff.Plan(rendered, store, isFullMode, graph)

	slog.Info("Publishing pages")
	renderedMap := makeRenderedMap(rendered)
	pub := publisher.New(client, cfg.Confluence.ParentPageID, spaceID, cfg.Publish.ApplyLabels)

	// Create hierarchy pages (component/version/module)
	hierarchy, err := pub.EnsureHierarchy(source)
	if err != nil {
		return fmt.Errorf("creating page hierarchy: %w", err)
	}

	// Set parent IDs for pages that don't have one yet
	for i := range plan.Items {
		if plan.Items[i].ParentID == "" && plan.Items[i].Action == model.ActionCreate {
			// Derive component/version/module from the page key
			// Key format: <siteKey>/<component>/<version>/<module>/<path>
			parts := strings.SplitN(plan.Items[i].Page.PageKey, "/", 5)
			if len(parts) >= 4 {
				parentID := hierarchy.ParentIDForPage(parts[1], parts[2], parts[3])
				if parentID != "" {
					plan.Items[i].ParentID = parentID
				}
			}
		}
	}

	reporter.PrintPlan(os.Stdout, plan)

	result := pub.Execute(plan, renderedMap)

	for _, pub := range result.PublishedPages {
		store.Upsert(state.Entry{
			PageKey:      pub.PageKey,
			ConfluenceID: pub.ConfluenceID,
			Fingerprint:  pub.Fingerprint,
			Title:        pub.Title,
			ParentID:     pub.ParentID,
			Version:      pub.Version,
		})
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
