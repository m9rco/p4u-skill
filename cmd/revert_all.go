package cmd

import (
	"fmt"

	"github.com/NZServer/NZMTools/p4u-skill/internal/ui"
	"github.com/spf13/cobra"
)

var revertAllCmd = &cobra.Command{
	Use:   "revert-all",
	Short: "Revert all opened files in all changelists",
	Long:  `Reverts all opened files across all pending changelists. Does not affect shelved files.`,
	RunE:  runRevertAll,
}

func init() {
	rootCmd.AddCommand(revertAllCmd)
}

func runRevertAll(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(cmd.OutOrStdout(), "Reverting all opened files...")
	if err := p4Client.RevertAll(); err != nil {
		return fmt.Errorf("revert-all: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), ui.Green.Sprint("✓ All files reverted."))
	return nil
}
