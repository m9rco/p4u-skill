package cmd

import (
	"fmt"

	"github.com/m9rco/p4u-skill/internal/p4"
	"github.com/m9rco/p4u-skill/internal/ui"
	"github.com/spf13/cobra"
)

var unshelveCmd = &cobra.Command{
	Use:   "unshelve [changelist]",
	Short: "Unshelve a changelist to itself (not the default changelist)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runUnshelve,
}

func init() {
	rootCmd.AddCommand(unshelveCmd)
}

func runUnshelve(cmd *cobra.Command, args []string) error {
	var cl string
	if len(args) == 1 {
		cl = args[0]
	} else {
		info, err := p4Client.GetInfo()
		if err != nil {
			return err
		}
		numbers, err := p4Client.ListChanges(p4.ListChangesOpts{
			Status: p4.StatusShelved, User: info.UserName, Client: info.ClientName,
		})
		if err != nil || len(numbers) == 0 {
			return fmt.Errorf("no shelved changelists found")
		}
		picked, err := ui.PickChangelist(numbers, nil, "unshelve", globalNonInteractive)
		if err != nil {
			return err
		}
		cl = picked
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Unshelving %s to itself...\n", ui.Cyan.Sprint(cl))
	if err := p4Client.Unshelve(cl, cl); err != nil {
		return fmt.Errorf("unshelve: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s Unshelved changelist %s\n", ui.Green.Sprint("✓"), cl)
	return nil
}
