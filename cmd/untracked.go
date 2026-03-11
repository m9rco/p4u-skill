package cmd

import (
	"fmt"

	"github.com/m9rco/p4u-skill/internal/output"
	"github.com/m9rco/p4u-skill/internal/ui"
	"github.com/spf13/cobra"
)

var untrackedDepth int

var untrackedCmd = &cobra.Command{
	Use:   "untracked [dir...]",
	Short: "Find files in p4 directories that are not tracked by Perforce",
	Long:  `Searches specified directories (default: current directory) for files not tracked in Perforce.`,
	RunE:  runUntracked,
}

func init() {
	untrackedCmd.Flags().IntVarP(&untrackedDepth, "depth", "d", 0, "Maximum directory depth (0 = unlimited)")
	rootCmd.AddCommand(untrackedCmd)
}

func runUntracked(cmd *cobra.Command, args []string) error {
	printer := output.New(globalJSON)

	dirs := args
	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "%s Searching for untracked files...\n", ui.Cyan.Sprint("p4u:"))

	files, err := p4Client.FindUntracked(dirs, untrackedDepth)
	if err != nil {
		return err
	}

	if printer.IsJSON() {
		printer.PrintJSON(map[string]interface{}{"untracked": files})
		return nil
	}

	for _, f := range files {
		fmt.Println(f)
	}
	if len(files) == 0 {
		fmt.Fprintln(cmd.ErrOrStderr(), ui.Green.Sprint("No untracked files found."))
	}
	return nil
}
