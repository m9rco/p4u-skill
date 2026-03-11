package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/m9rco/p4u-skill/internal/p4"
	"github.com/m9rco/p4u-skill/internal/ui"
	"github.com/spf13/cobra"
)

var (
	deleteClientForce   bool
	deleteClientName    string
	deleteClientNoFiles bool
)

var deleteClientCmd = &cobra.Command{
	Use:   "delete-client",
	Short: "Delete a p4 client completely (changelists, server record, and local files)",
	Long: `Deletes all changelists for the client, removes the client specification
from the server, and optionally deletes the local files.`,
	RunE: runDeleteClient,
}

func init() {
	deleteClientCmd.Flags().BoolVarP(&deleteClientForce, "force", "f", false, "Force delete without prompts")
	deleteClientCmd.Flags().StringVarP(&deleteClientName, "client", "c", "", "Client name to delete")
	deleteClientCmd.Flags().BoolVarP(&deleteClientNoFiles, "no-files", "n", false, "Don't delete local files")
	rootCmd.AddCommand(deleteClientCmd)
}

func runDeleteClient(cmd *cobra.Command, args []string) error {
	info, err := p4Client.GetInfo()
	if err != nil {
		return err
	}
	user := info.UserName

	clientName := deleteClientName
	clientPath := ""
	if clientName == "" {
		clientName = info.ClientName
		if clientName == "" || clientName == info.HostName {
			return fmt.Errorf("not inside a p4 client directory. Use -c to specify a client")
		}
		clientPath, _ = p4Client.GetClientPath(clientName)
	} else {
		// Look up full name and path.
		clients, err := p4Client.ListClients(user)
		if err != nil {
			return err
		}
		found := false
		for _, c := range clients {
			if c == clientName || containsSubstring(c, clientName) {
				clientName = c
				clientPath, _ = p4Client.GetClientPath(clientName)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("client %q not found or doesn't belong to user %s", deleteClientName, user)
		}
	}

	if deleteClientNoFiles || clientPath == "" {
		clientPath = ""
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Attempting to delete client: %s\n", ui.Cyan.Sprint(clientName))
	if clientPath != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "At path: %s\n", ui.Blue.Sprint(clientPath))
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), "[With no files to delete]")
	}

	if !deleteClientForce && !globalNonInteractive {
		if !ui.Prompt("Are you sure you want to proceed? [Y/n]", false) {
			fmt.Fprintln(cmd.OutOrStdout(), ui.Cyan.Sprint("Saved by the bell."))
			return nil
		}
	}

	// Get and delete all pending changelists.
	numbers, _ := p4Client.ListChanges(p4.ListChangesOpts{
		Status: p4.StatusPending, User: user, Client: clientName,
	})
	defFiles, _ := p4Client.DefaultOpenedFiles(user, clientName)

	if len(numbers) > 0 || len(defFiles) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "Found pending/shelved changelists: %v\n", numbers)
		if !deleteClientForce && !globalNonInteractive {
			if !ui.Prompt("Discard them? [Y/n]", false) {
				fmt.Fprintln(cmd.OutOrStdout(), ui.Cyan.Sprint("Saved by the bell."))
				return nil
			}
		}
		for _, cl := range numbers {
			clClient, _ := p4Client.GetChangeClient(cl)
			_ = p4Client.ShelveDelete(cl, clClient)
			files, _ := p4Client.OpenedInChangelist(cl, clClient)
			if len(files) > 0 {
				_ = p4Client.Revert(clClient, files)
			}
			fixes, _ := p4Client.GetFixes(cl)
			if len(fixes) > 0 {
				_ = p4Client.FixDelete(cl, fixes)
			}
			_ = p4Client.DeleteChange(cl, clClient)
		}
		if len(defFiles) > 0 {
			_ = p4Client.Revert(clientName, defFiles)
		}
	}

	// Unlock and delete client.
	fmt.Fprintln(cmd.OutOrStdout(), "Unlocking client...")
	if err := p4Client.UnlockClient(clientName); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not unlock: %v\n", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Deleting client from server...")
	if err := p4Client.DeleteClient(clientName); err != nil {
		return fmt.Errorf("delete client: %w", err)
	}

	// Delete local files.
	if clientPath != "" {
		fmt.Fprint(cmd.OutOrStdout(), "Deleting local files... ")
		if err := os.RemoveAll(clientPath); err != nil {
			return fmt.Errorf("remove %s: %w", clientPath, err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Done!")
		fmt.Fprintf(cmd.OutOrStdout(), "%s (%s is now gone)\n",
			ui.Cyan.Sprint("There is no spoon."), clientPath)
	}
	return nil
}

func containsSubstring(s, substr string) bool {
	return len(substr) > 0 && strings.Contains(s, substr)
}
