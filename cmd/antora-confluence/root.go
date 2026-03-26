package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	verbose   bool
	useDocker bool
	usePodman bool
)

var rootCmd = &cobra.Command{
	Use:   "antora-confluence",
	Short: "Publish Antora AsciiDoc documentation to Confluence",
	Long:  "A CLI tool that reads Antora-structured AsciiDoc repositories and publishes them as Confluence page hierarchies.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level := slog.LevelInfo
		if verbose {
			level = slog.LevelDebug
		}
		handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
		slog.SetDefault(slog.New(handler))
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (required)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&useDocker, "use-docker", false, "run asciidoctor inside Docker")
	rootCmd.PersistentFlags().BoolVar(&usePodman, "use-podman", false, "run asciidoctor inside Podman")
}
