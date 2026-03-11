package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Prompt asks the user a yes/no question and returns true for yes.
// If nonInteractive is true, it always returns true (assume yes).
func Prompt(question string, nonInteractive bool) bool {
	if nonInteractive {
		return true
	}
	fmt.Fprintf(os.Stderr, "%s ", question)
	reader := bufio.NewReader(os.Stdin)
	for {
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(strings.ToLower(line))
		switch line {
		case "y", "yes", "":
			return true
		case "n", "no":
			return false
		default:
			fmt.Fprint(os.Stderr, "Please answer yes or no: ")
		}
	}
}

// PickFromList shows a numbered list of items and asks the user to choose one.
// Returns the chosen item and its 0-based index.
// If nonInteractive is true and there's only one item, it's auto-selected.
func PickFromList(items []string, verb string, nonInteractive bool) (string, int, error) {
	if len(items) == 0 {
		return "", -1, fmt.Errorf("no items to pick from")
	}

	for i, item := range items {
		fmt.Fprintf(os.Stderr, "%s(%d) %s\n", Red.Sprint(""), i+1, item)
	}

	if len(items) == 1 {
		if nonInteractive {
			return items[0], 0, nil
		}
		question := Yellow.Sprintf("%s %s?", verb, items[0]) + " [Y/n]"
		if !Prompt(question, false) {
			return "", -1, fmt.Errorf("cancelled")
		}
		return items[0], 0, nil
	}

	if nonInteractive {
		return "", -1, fmt.Errorf("--non-interactive requires an explicit argument")
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "Which to %s? (1-%d): ", verb, len(items))
		line, _ := reader.ReadString('\n')
		n, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || n < 1 || n > len(items) {
			fmt.Fprintf(os.Stderr, "Invalid choice, enter a number between 1 and %d\n", len(items))
			continue
		}
		return items[n-1], n - 1, nil
	}
}
