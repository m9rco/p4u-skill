// Package cmd provides all p4u CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/m9rco/p4u-skill/internal/p4"
	"github.com/m9rco/p4u-skill/internal/ui"
	"github.com/spf13/cobra"
)

var (
	globalNoColor      bool
	globalForceColor   bool
	globalJSON         bool
	globalNonInteractive bool

	p4Client *p4.Client
)

var rootCmd = &cobra.Command{
	Use:   "p4u",
	Short: "A lighter approach to command-line Perforce (p4)",
	Long: `p4u is a cross-platform Perforce CLI enhancement tool.
It wraps common p4 workflows with better UX, color output, and automation support.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ui.InitColors(globalNoColor, globalForceColor)
		p4Client = p4.New()
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, ui.Red.Sprint("Error: ")+err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&globalNoColor, "no-color", "n", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVarP(&globalForceColor, "force-color", "o", false, "Force color output (when piping)")
	rootCmd.PersistentFlags().BoolVar(&globalJSON, "json", false, "Output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&globalNonInteractive, "non-interactive", false, "Disable interactive prompts (for automation)")
}
