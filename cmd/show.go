package cmd

import (
	"fmt"
	"strings"
	"sync"

	"github.com/NZServer/NZMTools/p4u-skill/internal/output"
	"github.com/NZServer/NZMTools/p4u-skill/internal/p4"
	"github.com/NZServer/NZMTools/p4u-skill/internal/ui"
	"github.com/spf13/cobra"
)

var (
	showVerbose  bool
	showBrief    bool
	showPending  bool
	showShelved  bool
	showAll      bool
	showDefault  bool
	showMax      int
	showClient   string
	showUser     string
	showNoLimit  bool
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current p4 client status and changelists",
	Long: `Prints the current pending and shelved changelists for the current p4 client.
Equivalent to the p4-show bash script.`,
	RunE: runShow,
}

func init() {
	showCmd.Flags().BoolVarP(&showVerbose, "verbose", "v", false, "Force verbose printing (show all filenames)")
	showCmd.Flags().BoolVarP(&showBrief, "brief", "b", false, "Force brief printing (no filenames)")
	showCmd.Flags().BoolVarP(&showPending, "pending", "p", false, "Show pending only (no shelved)")
	showCmd.Flags().BoolVarP(&showShelved, "shelved", "s", false, "Show shelved only")
	showCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all changelists")
	showCmd.Flags().BoolVarP(&showDefault, "default", "d", false, "Show only the default changelist")
	showCmd.Flags().IntVarP(&showMax, "max", "m", 0, "Show the N most recent changelists")
	showCmd.Flags().StringVarP(&showClient, "client", "c", "", "Filter by client")
	showCmd.Flags().StringVarP(&showUser, "user", "u", "", "Filter by user")
	showCmd.Flags().BoolVarP(&showNoLimit, "no-limit", "l", false, "Skip changelist count limit check")
	rootCmd.AddCommand(showCmd)
}

const changelistLimit = 50

func runShow(cmd *cobra.Command, args []string) error {
	printer := output.New(globalJSON)

	info, err := p4Client.GetInfo()
	if err != nil {
		return fmt.Errorf("p4 info failed: %w", err)
	}

	clientName := showClient
	if clientName == "" {
		clientName = info.ClientName
		if clientName == "" || clientName == info.HostName {
			return fmt.Errorf("not inside a p4 client directory. Use -c to specify a client")
		}
	}

	user := showUser
	if user == "" {
		user = info.UserName
	}

	// Determine which changelists to fetch.
	status := p4.StatusPending
	if showShelved {
		status = p4.StatusShelved
	} else if showAll || showMax > 0 {
		status = p4.StatusAll
	}

	// Show default changelist files.
	var defaultFiles []string
	if !showShelved && !showAll && showMax == 0 {
		defaultFiles, _ = p4Client.DefaultOpenedFiles(user, clientName)
	}

	if showDefault {
		// Only print default changelist.
		return printDefaultChangelist(printer, defaultFiles, showBrief)
	}

	opts := p4.ListChangesOpts{
		Status: status,
		User:   user,
		Client: clientName,
		Max:    showMax,
	}
	numbers, err := p4Client.ListChanges(opts)
	if err != nil {
		return err
	}

	// If pending only, subtract shelved.
	if showPending {
		shelvedNums, _ := p4Client.ListChanges(p4.ListChangesOpts{
			Status: p4.StatusShelved, User: user, Client: clientName,
		})
		shelvedSet := make(map[string]bool)
		for _, n := range shelvedNums {
			shelvedSet[n] = true
		}
		var pending []string
		for _, n := range numbers {
			if !shelvedSet[n] {
				pending = append(pending, n)
			}
		}
		numbers = pending
	}

	// Check limit.
	if !showNoLimit && len(numbers) > changelistLimit {
		fmt.Printf("There are %s changelists matching your criteria!\n",
			ui.Red.Sprintf("%d", len(numbers)))
		fmt.Println("Use -l to disable this check or -m to limit the shown changelists.")
		if !ui.Prompt("Continue?", globalNonInteractive) {
			return nil
		}
	}

	// Determine brief mode based on count.
	brief := showBrief
	if !showVerbose && !showBrief && len(numbers) > 4 {
		brief = true
	}
	if showVerbose {
		brief = false
	}

	if len(numbers) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "%s Fetching %d changelists...\n",
			ui.Cyan.Sprint("p4u:"), len(numbers))

		// Fetch all changelists concurrently.
		type result struct {
			idx    int
			output string
			err    error
		}
		results := make([]result, len(numbers))
		var wg sync.WaitGroup
		wg.Add(len(numbers))
		for i, num := range numbers {
			i, num := i, num
			go func() {
				defer wg.Done()
				out, e := formatChangelist(num, brief)
				results[i] = result{idx: i, output: out, err: e}
			}()
		}
		wg.Wait()

		if printer.IsJSON() {
			var jsonResults []interface{}
			for _, r := range results {
				if r.err == nil {
					jsonResults = append(jsonResults, r.output)
				}
			}
			printer.PrintJSON(jsonResults)
		} else {
			for _, r := range results {
				if r.err != nil {
					fmt.Fprintln(cmd.ErrOrStderr(), ui.Red.Sprint("Error: ")+r.err.Error())
				} else {
					fmt.Println(r.output)
				}
			}
		}
	}

	if len(defaultFiles) > 0 {
		return printDefaultChangelist(printer, defaultFiles, brief)
	}
	return nil
}

func printDefaultChangelist(printer *output.Printer, files []string, brief bool) error {
	if len(files) == 0 {
		return nil
	}
	if printer.IsJSON() {
		printer.PrintJSON(map[string]interface{}{
			"changelist": "default",
			"files":      files,
		})
		return nil
	}
	fmt.Println(ui.Purple.Sprint("Default changelist"))
	fmt.Println(ui.Blue.Sprint("    Pending Files"))
	if !brief {
		for _, f := range files {
			fmt.Println("   ", f)
		}
	}
	return nil
}

func formatChangelist(cl string, brief bool) (string, error) {
	d, err := p4Client.Describe(cl, false)
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	// Header line.
	header := fmt.Sprintf("Change %s", ui.Cyan.Sprint(cl))
	if d.Description != "" {
		firstLine := strings.SplitN(d.Description, "\n", 2)[0]
		header += " " + firstLine
	}
	if d.IsPending {
		header += " " + ui.Purple.Sprint("*pending*")
	}
	sb.WriteString(ui.Purple.Sprint(header) + "\n")

	if !brief {
		if d.Description != "" {
			sb.WriteString("    " + strings.ReplaceAll(d.Description, "\n", "\n    ") + "\n")
		}
		// Review links.
		for _, r := range d.ReviewLinks {
			sb.WriteString(ui.Yellow.Sprint("    "+r) + "\n")
		}
		// Bug fixes.
		if len(d.BugFixes) > 0 {
			sb.WriteString(ui.Green.Sprint("    Bugs Fixed: "))
			for _, b := range d.BugFixes {
				sb.WriteString(b + " ")
			}
			sb.WriteString("\n")
		}
	}

	// Fetch shelved info.
	shelved, _ := p4Client.Describe(cl, true)

	if len(d.PendingFiles) > 0 {
		sb.WriteString(ui.Blue.Sprint("    Pending Files") + "\n")
		if !brief {
			for _, f := range d.PendingFiles {
				sb.WriteString("        " + f + "\n")
			}
		}
	}
	if shelved != nil && len(shelved.ShelvedFiles) > 0 {
		sb.WriteString(ui.Red.Sprint("    Shelved Files") + "\n")
		if !brief {
			for _, f := range shelved.ShelvedFiles {
				sb.WriteString("        " + f + "\n")
			}
		}
	}

	return sb.String(), nil
}
