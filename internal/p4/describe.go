package p4

import (
	"fmt"
	"strings"
)

// ChangeDetail holds the detailed output of `p4 describe`.
type ChangeDetail struct {
	Number       string
	IsPending    bool
	Description  string
	User         string
	Client       string
	PendingFiles []string
	ShelvedFiles []string
	ReviewLinks  []string
	BugFixes     []string // raw lines for job fixes
}

// Describe fetches detailed information about a changelist.
// shelvedOnly=true uses -Ss (shelved info), false uses -s (pending info).
// brief=true passes -m1 so only one file per section is returned (faster for large CLs).
func (c *Client) Describe(cl string, shelvedOnly bool, brief ...bool) (*ChangeDetail, error) {
	flag := "-s"
	if shelvedOnly {
		flag = "-Ss"
	}
	args := []string{"describe", flag}
	if len(brief) > 0 && brief[0] {
		args = append(args, "-m1")
	}
	args = append(args, cl)
	out, err := c.exec.Run(args...)
	if err != nil {
		return nil, fmt.Errorf("describe %s: %w", cl, err)
	}
	return parseDescribe(cl, out), nil
}

func parseDescribe(cl, out string) *ChangeDetail {
	d := &ChangeDetail{Number: cl}
	lines := strings.Split(out, "\n")

	const (
		sectionAffected = "Affected files"
		sectionShelved  = "Shelved files"
		sectionJobs     = "Jobs fixed"
	)

	// State machine through the describe output.
	type section int
	const (
		secHeader section = iota
		secDescription
		secJobs
		secAffected
		secShelved
	)

	cur := secHeader
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch cur {
		case secHeader:
			// First non-empty line is the change header.
			if trimmed == "" {
				continue
			}
			if strings.Contains(line, "*pending*") {
				d.IsPending = true
			}
			cur = secDescription

		case secDescription:
			// Description lines are tab-indented; section headers are not.
			if strings.Contains(line, sectionJobs) {
				cur = secJobs
				continue
			}
			if strings.Contains(line, sectionAffected) {
				cur = secAffected
				continue
			}
			if strings.Contains(line, sectionShelved) {
				cur = secShelved
				continue
			}
			if trimmed == "" {
				continue
			}
			content := strings.TrimPrefix(line, "\t")
			// Detect ReviewBoard links in description.
			if strings.Contains(content, "ReviewBoard") || strings.Contains(content, "reviewboard") {
				d.ReviewLinks = append(d.ReviewLinks, strings.TrimSpace(content))
			}
			d.Description += content + "\n"

		case secJobs:
			if strings.Contains(line, sectionAffected) {
				cur = secAffected
				continue
			}
			if trimmed == "" {
				continue
			}
			d.BugFixes = append(d.BugFixes, trimmed)

		case secAffected:
			if strings.Contains(line, sectionShelved) {
				cur = secShelved
				continue
			}
			if trimmed == "" || trimmed == "..." {
				continue
			}
			// File lines start with "... //depot/..."
			content := strings.TrimPrefix(trimmed, "... ")
			d.PendingFiles = append(d.PendingFiles, content)

		case secShelved:
			if trimmed == "" || trimmed == "..." {
				continue
			}
			content := strings.TrimPrefix(trimmed, "... ")
			d.ShelvedFiles = append(d.ShelvedFiles, content)
		}
	}

	d.Description = strings.TrimSpace(d.Description)
	return d
}

// HasShelvedFiles returns true if the changelist has shelved content.
func (c *Client) HasShelvedFiles(cl string) (bool, error) {
	d, err := c.Describe(cl, true)
	if err != nil {
		return false, nil
	}
	return len(d.ShelvedFiles) > 0, nil
}

// HasPendingFiles returns true if the changelist has pending (open) files.
func (c *Client) HasPendingFiles(cl string) (bool, error) {
	d, err := c.Describe(cl, false)
	if err != nil {
		return false, nil
	}
	return len(d.PendingFiles) > 0, nil
}

// GetAnnotationCL returns the changelist number that last modified the given line
// of a depot file (using `p4 annotate -cq`).
func (c *Client) GetAnnotationCL(file string, line int) (string, error) {
	out, err := c.exec.Run("annotate", "-cq", file)
	if err != nil {
		return "", fmt.Errorf("annotate %s: %w", file, err)
	}
	lines := strings.Split(out, "\n")
	if line < 1 || line > len(lines) {
		return "", fmt.Errorf("line %d out of range (file has %d lines)", line, len(lines))
	}
	// Each annotate line looks like "12345: actual content"
	annotLine := lines[line-1]
	parts := strings.SplitN(annotLine, ":", 2)
	if len(parts) == 0 {
		return "", fmt.Errorf("could not parse annotation")
	}
	return strings.TrimSpace(parts[0]), nil
}
