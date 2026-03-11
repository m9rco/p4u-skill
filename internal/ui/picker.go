package ui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PickChangelist presents a list of changelists and prompts the user to choose.
// items is a list of "NNNN - description" strings.
// Returns the changelist number (just the first token) of the chosen item.
func PickChangelist(numbers []string, descriptions []string, verb string, nonInteractive bool) (string, error) {
	if len(numbers) == 0 {
		return "", fmt.Errorf("no changelists available")
	}

	for i, num := range numbers {
		desc := ""
		if i < len(descriptions) {
			desc = descriptions[i]
		}
		fmt.Fprintf(os.Stderr, "%s(%d) Change %s %s\n", Red.Sprint(""), i+1, Cyan.Sprint(num), desc)
	}

	if len(numbers) == 1 {
		if nonInteractive {
			return numbers[0], nil
		}
		q := Yellow.Sprintf("%s changelist %s?", verb, numbers[0]) + " [Y/n]"
		if !Prompt(q, false) {
			return "", fmt.Errorf("cancelled")
		}
		return numbers[0], nil
	}

	if nonInteractive {
		return "", fmt.Errorf("--non-interactive requires explicit changelist number")
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "Which changelist to %s? (1-%d): ", verb, len(numbers))
		line, _ := reader.ReadString('\n')
		n, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || n < 1 || n > len(numbers) {
			fmt.Fprintf(os.Stderr, "Invalid choice\n")
			continue
		}
		return numbers[n-1], nil
	}
}
