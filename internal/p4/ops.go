package p4

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// runWithStdin runs `p4 args...` with the given stdin content.
func runWithStdin(_ string, args ...string) error {
	// Last arg is the stdin content, actual p4 args are args[:len-1].
	if len(args) < 2 {
		return fmt.Errorf("runWithStdin: not enough args")
	}
	stdinContent := args[len(args)-1]
	p4Args := args[:len(args)-1]

	cmd := exec.Command("p4", p4Args...)
	cmd.Stdin = bytes.NewBufferString(stdinContent)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("p4 %s: %s", strings.Join(p4Args, " "), msg)
	}
	return nil
}

// UpdateClientHost changes the Host: field of a client spec using stdin/stdout.
func (c *Client) UpdateClientHost(clientName, newHost string) error {
	out, err := c.exec.Run("client", "-o", clientName)
	if err != nil {
		return fmt.Errorf("client -o %s: %w", clientName, err)
	}
	var updated strings.Builder
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Host:") {
			updated.WriteString("Host:\t" + newHost + "\n")
		} else {
			updated.WriteString(line + "\n")
		}
	}
	return runWithStdin(clientName, "client", "-i", updated.String())
}

// UnlockClient removes the "locked" option from a client spec.
func (c *Client) UnlockClient(clientName string) error {
	out, err := c.exec.Run("client", "-o", clientName)
	if err != nil {
		return fmt.Errorf("client -o %s: %w", clientName, err)
	}
	updated := strings.ReplaceAll(out, " locked ", " unlocked ")
	return runWithStdin(clientName, "client", "-i", updated)
}

// DeleteClient deletes a client from the server.
func (c *Client) DeleteClient(clientName string) error {
	_, err := c.exec.Run("client", "-d", clientName)
	return err
}

// LoginStatus checks whether the user has a valid p4 ticket.
func (c *Client) LoginStatus() error {
	_, err := c.exec.Run("login", "-s")
	return err
}

// ShelveDelete deletes the shelve of a changelist.
func (c *Client) ShelveDelete(cl, client string) error {
	args := []string{"shelve", "-d", "-c", cl}
	if client != "" {
		args = append([]string{"-c", client}, args...)
		// p4 -c client shelve -d -c CL
		args = append([]string{}, "shelve", "-d", "-c", cl)
		cmd := exec.Command("p4", append([]string{"-c", client}, args...)...)
		return cmd.Run()
	}
	_, err := c.exec.Run(args...)
	return err
}

// ShelveCreate shelves the pending files of a changelist.
func (c *Client) ShelveCreate(cl string) error {
	_, err := c.exec.Run("shelve", "-c", cl)
	return err
}

// Unshelve unshelves changelist src into changelist dst.
func (c *Client) Unshelve(src, dst string) error {
	_, err := c.exec.Run("unshelve", "-s", src, "-c", dst)
	return err
}

// Revert reverts a list of depot files.
func (c *Client) Revert(client string, files []string) error {
	if len(files) == 0 {
		return nil
	}
	args := []string{"revert"}
	if client != "" {
		args = append([]string{"-c", client}, args...)
		// full: p4 -c client revert files...
		cmd := exec.Command("p4", append(append([]string{"-c", client}, "revert"), files...)...)
		return cmd.Run()
	}
	args = append(args, files...)
	_, err := c.exec.Run(args...)
	return err
}

// RevertAll reverts all opened files.
func (c *Client) RevertAll() error {
	out, err := c.exec.Run("opened")
	if err != nil || strings.TrimSpace(out) == "" {
		return nil
	}
	var files []string
	for _, line := range strings.Split(out, "\n") {
		if idx := strings.Index(line, "#"); idx >= 0 {
			files = append(files, strings.TrimSpace(line[:idx]))
		}
	}
	if len(files) == 0 {
		return nil
	}
	args := append([]string{"revert"}, files...)
	_, err = c.exec.Run(args...)
	return err
}

// DeleteChange deletes a changelist.
func (c *Client) DeleteChange(cl, client string) error {
	if client != "" {
		cmd := exec.Command("p4", "-c", client, "change", "-d", cl)
		return cmd.Run()
	}
	_, err := c.exec.Run("change", "-d", cl)
	return err
}

// FixDelete removes a job fix from a changelist.
func (c *Client) FixDelete(cl string, jobs []string) error {
	if len(jobs) == 0 {
		return nil
	}
	args := append([]string{"fix", "-d", "-c", cl}, jobs...)
	_, err := c.exec.Run(args...)
	return err
}

// Sync runs `p4 sync`.
func (c *Client) Sync() error {
	_, err := c.exec.Run("sync")
	return err
}

// ResolveAutoMerge runs `p4 resolve -am`.
func (c *Client) ResolveAutoMerge() error {
	_, err := c.exec.Run("resolve", "-am")
	return err
}

// Resolve runs interactive `p4 resolve` by delegating to the system.
func (c *Client) Resolve() error {
	cmd := exec.Command("p4", "resolve")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// FindUntracked returns files in dirs that p4 does not know about.
func (c *Client) FindUntracked(dirs []string, maxDepth int) ([]string, error) {
	if len(dirs) == 0 {
		dirs = []string{"."}
	}
	// Build find command args.
	findArgs := dirs
	if maxDepth > 0 {
		findArgs = append(findArgs, "-maxdepth", fmt.Sprintf("%d", maxDepth))
	}
	findArgs = append(findArgs, "-type", "f")

	// Run find.
	findCmd := exec.Command("find", findArgs...)
	findOut, _ := findCmd.Output()
	allFiles := strings.Split(strings.TrimSpace(string(findOut)), "\n")

	var untracked []string
	for _, f := range allFiles {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		out, _ := c.exec.Run("fstat", f)
		if strings.Contains(out, "no such file") || out == "" {
			untracked = append(untracked, f)
		}
	}
	return untracked, nil
}
