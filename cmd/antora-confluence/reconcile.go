package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bovinemagnet/antora2confluence/internal/config"
	"github.com/bovinemagnet/antora2confluence/internal/confluence"
	"github.com/bovinemagnet/antora2confluence/internal/state"
	"github.com/spf13/cobra"
)

var reconcileCmd = &cobra.Command{
	Use:   "reconcile-state",
	Short: "Rebuild local state from Confluence page properties",
	Long:  "Walks managed pages under the configured root page in Confluence and rebuilds the local state file from page properties. Useful when local state is lost or corrupted.",
	RunE:  runReconcile,
}

func init() {
	rootCmd.AddCommand(reconcileCmd)
}

func runReconcile(cmd *cobra.Command, args []string) error {
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

	slog.Info("Walking Confluence pages under root", "rootPageID", cfg.Confluence.ParentPageID)
	store := state.New(cfg.Sync.StateFile)

	count, err := reconcilePages(client, store, cfg.Confluence.ParentPageID)
	if err != nil {
		return fmt.Errorf("reconciling state: %w", err)
	}

	if err := store.Save(); err != nil {
		return fmt.Errorf("saving state: %w", err)
	}

	slog.Info("Reconciliation complete", "pagesFound", count)
	fmt.Fprintf(os.Stdout, "\nReconciled %d managed pages into %s\n", count, cfg.Sync.StateFile)
	return nil
}

func reconcilePages(client *confluence.Client, store *state.Store, parentID string) (int, error) {
	children, err := client.GetChildPages(parentID)
	if err != nil {
		return 0, fmt.Errorf("getting children of %s: %w", parentID, err)
	}

	count := 0
	for _, child := range children {
		prop, err := client.GetPageProperty(child.ID, "antora-page-key")
		if err != nil {
			slog.Warn("Failed to read property", "pageID", child.ID, "error", err)
			continue
		}

		if prop != nil {
			pageKey, ok := prop.Value.(string)
			if ok && pageKey != "" {
				fpProp, _ := client.GetPageProperty(child.ID, "antora-fingerprint")
				fingerprint := ""
				if fpProp != nil {
					if fp, ok := fpProp.Value.(string); ok {
						fingerprint = fp
					}
				}

				version := 0
				if child.Version != nil {
					version = child.Version.Number
				}

				store.Upsert(state.Entry{
					PageKey:      pageKey,
					ConfluenceID: child.ID,
					Title:        child.Title,
					Fingerprint:  fingerprint,
					ParentID:     parentID,
					Version:      version,
				})
				count++
				slog.Debug("Found managed page", "pageKey", pageKey, "id", child.ID)
			}
		}

		childCount, err := reconcilePages(client, store, child.ID)
		if err != nil {
			return count, err
		}
		count += childCount
	}

	return count, nil
}
