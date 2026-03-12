package cmd

import (
	"fmt"

	"github.com/m9rco/p4u-skill/internal/p4"
	"github.com/m9rco/p4u-skill/internal/ui"
	"github.com/spf13/cobra"
)

var deleteCLForce bool

var deleteCLCmd = &cobra.Command{
	Use:   "delete-cl [changelist...]",
	Short: "Delete a changelist completely (shelve, pending files, fixes)",
	Long: `Removes the shelve, reverts all pending files, removes job fixes,
and then deletes the changelist. Equivalent to p4-delete-changelist.`,
	RunE: runDeleteCL,
}

func init() {
	deleteCLCmd.Flags().BoolVarP(&deleteCLForce, "force", "f", false, "Force delete without prompts")
	rootCmd.AddCommand(deleteCLCmd)
}

func runDeleteCL(cmd *cobra.Command, args []string) error {
	var changelists []string
	if len(args) > 0 {
		changelists = args
	} else {
		info, err := p4Client.GetInfo()
		if err != nil {
			return err
		}
		numbers, err := p4Client.ListChanges(p4.ListChangesOpts{
			Status: p4.StatusPending, User: info.UserName, Client: info.ClientName,
		})
		if err != nil || len(numbers) == 0 {
			return fmt.Errorf("no changelists to delete")
		}
		picked, err := ui.PickChangelist(numbers, nil, "delete", globalNonInteractive || deleteCLForce)
		if err != nil {
			return err
		}
		changelists = []string{picked}
	}

	for _, cl := range changelists {
		// Verify CL exists.
		if _, err := p4Client.Describe(cl, false); err != nil {
			return fmt.Errorf("changelist %s not found: %w", cl, err)
		}

		// Get the client associated with this CL.
		clClient, err := p4Client.GetChangeClient(cl)
		if err != nil {
			clClient = ""
		}

		// Ensure client hostname matches current machine before operations.
		_ = p4Client.EnsureHostname(clClient)

		fmt.Fprintf(cmd.OutOrStdout(), "Deleting changelist %s...\n", ui.Cyan.Sprint(cl))

		// Delete shelve.
		_ = p4Client.ShelveDelete(cl, clClient)

		// Revert pending files.
		files, _ := p4Client.OpenedInChangelist(cl, clClient)
		if len(files) > 0 {
			if err := p4Client.Revert(clClient, files); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: revert failed: %v\n", err)
			}
		}

		// Remove job fixes.
		fixes, _ := p4Client.GetFixes(cl)
		if len(fixes) > 0 {
			_ = p4Client.FixDelete(cl, fixes)
		}

		// Delete the changelist.
		if err := p4Client.DeleteChange(cl, clClient); err != nil {
			return fmt.Errorf("delete changelist %s: %w", cl, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s Changelist %s deleted.\n",
			ui.Green.Sprint("✓"), cl)
	}
	return nil
}
