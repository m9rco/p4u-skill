package cmd

import (
	"fmt"
	"strings"

	"github.com/m9rco/p4u-skill/internal/output"
	"github.com/m9rco/p4u-skill/internal/p4"
	"github.com/m9rco/p4u-skill/internal/ui"
	"github.com/spf13/cobra"
)

var showCLBrief bool

var showCLCmd = &cobra.Command{
	Use:   "show-cl [changelist]",
	Short: "Pretty-print a single changelist",
	Long: `Prints detailed information about a specific changelist including
pending files, shelved files, review links, and bug fixes.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShowCL,
}

func init() {
	showCLCmd.Flags().BoolVarP(&showCLBrief, "brief", "b", false, "Brief output (no file list)")
	rootCmd.AddCommand(showCLCmd)
}

func runShowCL(cmd *cobra.Command, args []string) error {
	printer := output.New(globalJSON)

	var cl string
	if len(args) == 1 {
		cl = args[0]
	} else {
		// Interactive pick.
		info, err := p4Client.GetInfo()
		if err != nil {
			return fmt.Errorf("p4 info: %w", err)
		}
		numbers, err := p4Client.ListChanges(p4.ListChangesOpts{
			Status: p4.StatusPending, User: info.UserName, Client: info.ClientName,
		})
		if err != nil || len(numbers) == 0 {
			return fmt.Errorf("no changelists found")
		}
		picked, err := ui.PickChangelist(numbers, nil, "show", globalNonInteractive)
		if err != nil {
			return err
		}
		cl = picked
	}

	d, err := p4Client.Describe(cl, false)
	if err != nil {
		return err
	}
	shelved, _ := p4Client.Describe(cl, true)

	if printer.IsJSON() {
		type jsonOut struct {
			Number      string   `json:"number"`
			IsPending   bool     `json:"is_pending"`
			Description string   `json:"description"`
			ReviewLinks []string `json:"review_links,omitempty"`
			BugFixes    []string `json:"bug_fixes,omitempty"`
			Pending     []string `json:"pending_files,omitempty"`
			Shelved     []string `json:"shelved_files,omitempty"`
		}
		out := jsonOut{
			Number:      cl,
			IsPending:   d.IsPending,
			Description: d.Description,
			ReviewLinks: d.ReviewLinks,
			BugFixes:    d.BugFixes,
			Pending:     d.PendingFiles,
		}
		if shelved != nil {
			out.Shelved = shelved.ShelvedFiles
		}
		printer.PrintJSON(out)
		return nil
	}

	// Text output.
	header := fmt.Sprintf("Change %s", ui.Cyan.Sprint(cl))
	if d.Description != "" {
		firstLine := strings.SplitN(d.Description, "\n", 2)[0]
		header += " " + firstLine
	}
	if d.IsPending {
		header += " " + ui.Purple.Sprint("*pending*")
	}
	fmt.Println(ui.Purple.Sprint(header))

	if !showCLBrief {
		if d.Description != "" {
			fmt.Println("    " + strings.ReplaceAll(d.Description, "\n", "\n    "))
		}
		for _, r := range d.ReviewLinks {
			fmt.Println(ui.Yellow.Sprint("    " + r))
		}
		if len(d.BugFixes) > 0 {
			fmt.Print(ui.Green.Sprint("    Bugs Fixed: "))
			fmt.Println(strings.Join(d.BugFixes, " "))
		}
	}

	if len(d.PendingFiles) > 0 {
		fmt.Println(ui.Blue.Sprint("    Pending Files"))
		if !showCLBrief {
			for _, f := range d.PendingFiles {
				fmt.Println("        " + f)
			}
		}
	}
	if shelved != nil && len(shelved.ShelvedFiles) > 0 {
		fmt.Println(ui.Red.Sprint("    Shelved Files"))
		if !showCLBrief {
			for _, f := range shelved.ShelvedFiles {
				fmt.Println("        " + f)
			}
		}
	}
	return nil
}
