package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bovinemagnet/antora2confluence/internal/config"
	"github.com/bovinemagnet/antora2confluence/internal/confluence"
	"github.com/bovinemagnet/antora2confluence/internal/depgraph"
	"github.com/bovinemagnet/antora2confluence/internal/diff"
	"github.com/bovinemagnet/antora2confluence/internal/discovery"
	"github.com/bovinemagnet/antora2confluence/internal/renderer"
	"github.com/bovinemagnet/antora2confluence/internal/reporter"
	"github.com/bovinemagnet/antora2confluence/internal/state"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Show what would be published without making changes",
	Long:  "Performs discovery, rendering, and change detection, then displays the planned actions without executing them. Requires Confluence access to validate auth and resolve space ID.",
	RunE:  runPlan,
}

func init() {
	rootCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
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

	slog.Info("Validating Confluence access")
	client := confluence.NewClient(cfg.Confluence.BaseURL, username, token)
	if _, err := client.ValidateAuth(cfg.Confluence.SpaceKey); err != nil {
		return fmt.Errorf("validating Confluence access: %w", err)
	}

	slog.Info("Scanning Antora content", "root", cfg.Source.AntoraRoot)
	source, err := discovery.Scan(cfg.Source.AntoraRoot, cfg.Source.SiteKey)
	if err != nil {
		return fmt.Errorf("scanning Antora source: %w", err)
	}
	pages := source.AllPages()
	slog.Info("Discovered pages", "count", len(pages))

	slog.Info("Rendering pages")
	bc := buildBackendConfig(cfg, cfg.Source.AntoraRoot)
	r := renderer.New(bc, cfg.Render.MermaidMode)
	rendered, renderErrs := r.RenderAll(pages)
	if len(renderErrs) > 0 {
		slog.Warn("Some pages failed to render", "failures", len(renderErrs))
	}

	graph := depgraph.Build(rendered)

	store := state.New(cfg.Sync.StateFile)
	if err := store.Load(); err != nil {
		return fmt.Errorf("loading state: %w", err)
	}

	plan := diff.Plan(rendered, store, false, graph)
	reporter.PrintPlan(os.Stdout, plan)
	fmt.Fprintf(os.Stdout, "\nDry run — no changes made.\n")
	return nil
}
