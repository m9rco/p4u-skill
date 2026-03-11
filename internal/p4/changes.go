package p4

import (
	"fmt"
	"strings"
)

// ChangeStatus is the status of a changelist.
type ChangeStatus string

const (
	StatusPending   ChangeStatus = "pending"
	StatusShelved   ChangeStatus = "shelved"
	StatusSubmitted ChangeStatus = "submitted"
	StatusAll       ChangeStatus = ""
)

// Change represents a Perforce changelist entry (from `p4 changes`).
type Change struct {
	Number      string
	User        string
	Client      string
	Description string
	Status      ChangeStatus
}

// ListChangesOpts configures `p4 changes` queries.
type ListChangesOpts struct {
	Status ChangeStatus
	User   string
	Client string
	Max    int
}

// ListChanges runs `p4 changes` and returns the changelist numbers.
func (c *Client) ListChanges(opts ListChangesOpts) ([]string, error) {
	args := []string{"changes"}
	if opts.Status != StatusAll {
		args = append(args, "-s", string(opts.Status))
	}
	if opts.User != "" {
		args = append(args, "-u", opts.User)
	}
	if opts.Client != "" {
		args = append(args, "-c", opts.Client)
	}
	if opts.Max > 0 {
		args = append(args, "-m", fmt.Sprintf("%d", opts.Max))
	}

	out, err := c.exec.Run(args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var numbers []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// format: "Change 12345 on 2024/01/01 by user@client 'description'"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			numbers = append(numbers, parts[1])
		}
	}
	return numbers, nil
}

// DefaultOpenedFiles returns the depot files opened in the default changelist.
func (c *Client) DefaultOpenedFiles(user, client string) ([]string, error) {
	args := []string{"opened"}
	if user != "" {
		args = append(args, "-u", user)
	}
	if client != "" {
		args = append(args, "-C", client)
	}
	out, err := c.exec.Run(args...)
	if err != nil {
		return nil, nil // no opened files is not an error
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "default change") {
			depotFile := line
			if idx := strings.Index(depotFile, "#"); idx >= 0 {
				depotFile = depotFile[:idx]
			}
			files = append(files, strings.TrimSpace(depotFile))
		}
	}
	return files, nil
}
