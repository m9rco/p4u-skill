package cmd

import (
	"fmt"
	"strings"

	"github.com/m9rco/p4u-skill/internal/p4"
	"github.com/m9rco/p4u-skill/internal/ui"
	"github.com/spf13/cobra"
)

var (
	switchPick        bool
	switchKeepShelved bool
	switchSuppressDef bool
	switchVerbose     bool
	switchSync        bool
	switchResolve     bool
	switchAutoMerge   bool
)

var switchCmd = &cobra.Command{
	Use:   "switch [changelist...]",
	Short: "Switch working context by shelving current work and unshelving target",
	Long: `Shelves all current pending changelists and unshelves the specified
changelist(s). Run without arguments to shelve everything and start fresh.`,
	RunE: runSwitch,
}

func init() {
	switchCmd.Flags().BoolVarP(&switchPick, "pick", "p", false, "Interactively pick a changelist to switch to")
	switchCmd.Flags().BoolVarP(&switchKeepShelved, "keep", "k", false, "Keep shelves after unshelving")
	switchCmd.Flags().BoolVarP(&switchSuppressDef, "suppress-default", "d", false, "Suppress default changelist warning")
	switchCmd.Flags().BoolVarP(&switchVerbose, "verbose", "v", false, "Verbose output")
	switchCmd.Flags().BoolVarP(&switchSync, "sync", "s", false, "Sync after unshelve")
	switchCmd.Flags().BoolVarP(&switchResolve, "resolve", "r", false, "Resolve all conflicts after sync")
	switchCmd.Flags().BoolVarP(&switchAutoMerge, "auto-merge", "m", false, "Auto-resolve merge conflicts only")
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	if err := p4Client.LoginStatus(); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	info, err := p4Client.GetInfo()
	if err != nil {
		return err
	}
	clientName := info.ClientName
	if clientName == "" || clientName == info.HostName {
		return fmt.Errorf("not inside a p4 client directory")
	}
	user := info.UserName

	// Resolve target changelists.
	var targetCLs []string
	if len(args) > 0 {
		targetCLs = args
	}
	if switchPick {
		numbers, err := p4Client.ListChanges(p4.ListChangesOpts{
			Status: p4.StatusPending, User: user, Client: clientName,
		})
		if err != nil || len(numbers) == 0 {
			return fmt.Errorf("no changelists to pick from")
		}
		picked, err := ui.PickChangelist(numbers, nil, "switch context to", globalNonInteractive)
		if err != nil {
			return err
		}
		targetCLs = []string{picked}
	}

	// Check for default changelist.
	defFiles, _ := p4Client.DefaultOpenedFiles(user, clientName)
	if !switchSuppressDef && len(defFiles) > 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), ui.Yellow.Sprint(
			"Warning: Pending default changelist detected - overlapping files may be overridden."))
		if !ui.Prompt("Do you wish to continue? [Y/n]", globalNonInteractive) {
			return nil
		}
	}

	// Get all pending changelists.
	allCLs, err := p4Client.ListChanges(p4.ListChangesOpts{
		Status: p4.StatusPending, User: user, Client: clientName,
	})
	if err != nil {
		return err
	}
	if len(allCLs) == 0 {
		return fmt.Errorf("no open changelists")
	}

	targetSet := make(map[string]bool)
	for _, cl := range targetCLs {
		targetSet[cl] = true
	}

	// Shelve all non-target changelists.
	for _, cl := range allCLs {
		if targetSet[cl] {
			continue
		}
		hasPending, _ := p4Client.HasPendingFiles(cl)
		if !hasPending {
			logSwitchVerbose(cmd, switchVerbose, "Changelist %s - unchanged (shelved)\n", cl)
			continue
		}
		isShelved, _ := p4Client.HasShelvedFiles(cl)
		if isShelved {
			fmt.Fprintf(cmd.ErrOrStderr(), "Changelist %s already shelved.\n", cl)
			if !ui.Prompt("Delete current shelve and re-shelve? [Y/n]", globalNonInteractive) {
				return fmt.Errorf("aborted")
			}
			logSwitchVerbose(cmd, switchVerbose, "Changelist %s - reshelve & revert\n", cl)
			_ = p4Client.ShelveDelete(cl, clientName)
		} else {
			logSwitchVerbose(cmd, switchVerbose, "Changelist %s - shelve & revert\n", cl)
		}
		if err := p4Client.ShelveCreate(cl); err != nil {
			return fmt.Errorf("shelve %s: %w", cl, err)
		}
		files, _ := p4Client.OpenedInChangelist(cl, clientName)
		if len(files) > 0 {
			_ = p4Client.Revert(clientName, files)
		}
	}

	// Unshelve target changelists.
	for _, cl := range targetCLs {
		isShelved, _ := p4Client.HasShelvedFiles(cl)
		if !isShelved {
			logSwitchVerbose(cmd, switchVerbose, "Changelist %s - unchanged (pending)\n", cl)
			continue
		}
		hasPending, _ := p4Client.HasPendingFiles(cl)
		if hasPending {
			logSwitchVerbose(cmd, switchVerbose, "Changelist %s - unchanged (has pending files)\n", cl)
			continue
		}
		keepMsg := ""
		if switchKeepShelved {
			keepMsg = " (keeping shelve)"
		}
		logSwitchVerbose(cmd, switchVerbose, "Changelist %s - unshelve%s\n", cl, keepMsg)
		if err := p4Client.Unshelve(cl, cl); err != nil {
			return fmt.Errorf("unshelve %s: %w", cl, err)
		}
		if !switchKeepShelved {
			_ = p4Client.ShelveDelete(cl, "")
		}
	}

	if switchSync {
		fmt.Fprintln(cmd.OutOrStdout(), ui.Cyan.Sprint("Syncing..."))
		if err := p4Client.Sync(); err != nil {
			return fmt.Errorf("sync: %w", err)
		}
	}

	if switchAutoMerge {
		fmt.Fprintln(cmd.OutOrStdout(), ui.Cyan.Sprint("Resolving auto merge..."))
		_ = p4Client.ResolveAutoMerge()
		if switchResolve {
			fmt.Fprintln(cmd.OutOrStdout(), ui.Cyan.Sprint("Resolving manual..."))
			_ = p4Client.Resolve()
		}
	} else if switchResolve {
		fmt.Fprintln(cmd.OutOrStdout(), ui.Cyan.Sprint("Resolving..."))
		_ = p4Client.Resolve()
	}

	return nil
}

func logSwitchVerbose(cmd *cobra.Command, verbose bool, format string, a ...interface{}) {
	if verbose {
		msg := fmt.Sprintf(format, a...)
		fmt.Fprint(cmd.OutOrStdout(), ui.Cyan.Sprint(strings.TrimRight(msg, "\n"))+"\n")
	}
}
