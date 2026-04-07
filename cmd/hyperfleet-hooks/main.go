package main

import (
	"os"
	"path/filepath"

	"github.com/openshift-hyperfleet/hyperfleet-hooks/cmd/hyperfleet-hooks/commitlint"
	"github.com/openshift-hyperfleet/hyperfleet-hooks/cmd/hyperfleet-hooks/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "hyperfleet-hooks",
	Short:        "HyperFleet pre-commit hooks validator",
	Long:         "hyperfleet-hooks validates artifacts against HyperFleet project standards.",
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(version.NewCommand())
	rootCmd.AddCommand(commitlint.NewCommand())
}

func main() {
	// Auto-detect if called via symlink (for pre-commit compatibility)
	execName := filepath.Base(os.Args[0])
	if execName == "hyperfleet-commitlint" {
		// Called as hyperfleet-commitlint, auto-run commitlint subcommand
		os.Args = append([]string{"hyperfleet-hooks", "commitlint"}, os.Args[1:]...)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
