package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/m9rco/p4u-skill/internal/output"
	"github.com/m9rco/p4u-skill/internal/ui"
	"github.com/spf13/cobra"
)

var annotateVerbose bool

var annotateCmd = &cobra.Command{
	Use:   "annotate <file> <line>",
	Short: "Show the changelist that last modified a specific line in a file",
	Long: `Annotates a specific line in a depot file, showing which changelist
last modified it. Also follows copy/merge history.`,
	Args: cobra.ExactArgs(2),
	RunE: runAnnotate,
}

func init() {
	annotateCmd.Flags().BoolVarP(&annotateVerbose, "verbose", "v", false, "Verbose changelist output (show file list)")
	rootCmd.AddCommand(annotateCmd)
}

// innerCLPattern matches references like "CL 12345", "changelist 12345", "copy from 12345", etc.
var innerCLPattern = regexp.MustCompile(`(?i)(?:cl|changelist|copy from|edit from|merge from)[,\s-]+([0-9]+)`)

func runAnnotate(cmd *cobra.Command, args []string) error {
	file := args[0]
	printer := output.New(globalJSON)

	var lineNum int
	if _, err := fmt.Sscanf(args[1], "%d", &lineNum); err != nil {
		return fmt.Errorf("invalid line number %q", args[1])
	}

	cl, err := p4Client.GetAnnotationCL(file, lineNum)
	if err != nil {
		return err
	}
	if cl == "" {
		return fmt.Errorf("annotation not found for line %d", lineNum)
	}

	var results []interface{}
	visited := make(map[string]bool)

	for cl != "" && !visited[cl] {
		visited[cl] = true

		d, err := p4Client.Describe(cl, false)
		if err != nil {
			break
		}

		if printer.IsJSON() {
			type jsonCL struct {
				Number      string   `json:"number"`
				Description string   `json:"description"`
				User        string   `json:"user"`
				ReviewLinks []string `json:"review_links,omitempty"`
				BugFixes    []string `json:"bug_fixes,omitempty"`
				Files       []string `json:"files,omitempty"`
			}
			entry := jsonCL{
				Number:      cl,
				Description: d.Description,
				ReviewLinks: d.ReviewLinks,
				BugFixes:    d.BugFixes,
			}
			if annotateVerbose {
				entry.Files = d.PendingFiles
			}
			results = append(results, entry)
		} else {
			fmt.Println(ui.Purple.Sprintf("Change %s", cl))
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
			if annotateVerbose && len(d.PendingFiles) > 0 {
				fmt.Println(ui.Blue.Sprint("    Files:"))
				for _, f := range d.PendingFiles {
					fmt.Println("        " + f)
				}
			}
			fmt.Println()
		}

		// Look for inner CL references in the description.
		cl = ""
		m := innerCLPattern.FindStringSubmatch(d.Description)
		if len(m) >= 2 {
			candidate := m[1]
			if !visited[candidate] {
				cl = candidate
			}
		}
	}

	if printer.IsJSON() {
		printer.PrintJSON(results)
	}
	return nil
}
