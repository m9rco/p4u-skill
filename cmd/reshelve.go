package cmd

import (
	"fmt"

	"github.com/NZServer/NZMTools/p4u-skill/internal/p4"
	"github.com/NZServer/NZMTools/p4u-skill/internal/ui"
	"github.com/spf13/cobra"
)

var reshelveCmd = &cobra.Command{
	Use:   "reshelve [changelist]",
	Short: "Re-shelve a changelist (delete old shelve and shelve current pending)",
	Long:  `Deletes the existing shelve of a changelist and re-creates it from the current pending files.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runReshelve,
}

func init() {
	rootCmd.AddCommand(reshelveCmd)
}

func runReshelve(cmd *cobra.Command, args []string) error {
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
		picked, err := ui.PickChangelist(numbers, nil, "reshelve", globalNonInteractive)
		if err != nil {
			return err
		}
		cl = picked
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Deleting existing shelve for %s...\n", ui.Cyan.Sprint(cl))
	if err := p4Client.ShelveDelete(cl, ""); err != nil {
		return fmt.Errorf("delete shelve: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Re-shelving %s...\n", ui.Cyan.Sprint(cl))
	if err := p4Client.ShelveCreate(cl); err != nil {
		return fmt.Errorf("shelve: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s Reshelved changelist %s\n", ui.Green.Sprint("✓"), cl)
	return nil
}
