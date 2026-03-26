package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bovinemagnet/antora2confluence/internal/config"
	"github.com/bovinemagnet/antora2confluence/internal/discovery"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate Antora source structure without publishing",
	Long:  "Scans the Antora content root, discovers components, modules, versions, and pages, and reports the content inventory. Does not connect to Confluence.",
	RunE:  runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	if cfgFile == "" {
		return fmt.Errorf("--config flag is required")
	}

	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	slog.Info("Scanning Antora content", "root", cfg.Source.AntoraRoot, "siteKey", cfg.Source.SiteKey)

	source, err := discovery.Scan(cfg.Source.AntoraRoot, cfg.Source.SiteKey)
	if err != nil {
		return fmt.Errorf("scanning Antora source: %w", err)
	}

	pages := source.AllPages()
	fmt.Fprintf(os.Stdout, "\nContent Inventory\n")
	fmt.Fprintf(os.Stdout, "=================\n\n")

	for _, comp := range source.Components {
		fmt.Fprintf(os.Stdout, "Component: %s\n", comp.Name)
		for _, ver := range comp.Versions {
			fmt.Fprintf(os.Stdout, "  Version: %s\n", ver.Name)
			for _, mod := range ver.Modules {
				fmt.Fprintf(os.Stdout, "    Module: %s (%d pages, %d nav entries)\n",
					mod.Name, len(mod.Pages), len(mod.Nav))
				for _, page := range mod.Pages {
					key := page.PageKey
					if page.ExplicitKey != "" {
						key = page.ExplicitKey + " (explicit)"
					}
					fmt.Fprintf(os.Stdout, "      Page: %s [%s]\n", page.Title, key)
					if len(page.Images) > 0 {
						fmt.Fprintf(os.Stdout, "            Images: %v\n", page.Images)
					}
					if len(page.XRefs) > 0 {
						fmt.Fprintf(os.Stdout, "            XRefs: %v\n", page.XRefs)
					}
					if len(page.Includes) > 0 {
						fmt.Fprintf(os.Stdout, "            Includes: %v\n", page.Includes)
					}
				}
			}
		}
	}

	fmt.Fprintf(os.Stdout, "\nTotal: %d components, %d pages\n",
		len(source.Components), len(pages))
	fmt.Fprintf(os.Stdout, "Validation passed.\n")

	return nil
}
