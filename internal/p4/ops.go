package p4

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// EnsureHostname checks if the client's Host field matches the current machine.
// If it differs, updates the client spec so operations aren't rejected by the server.
// Mirrors the shell p4-fix-hostname helper. Silently no-ops if host field is empty.
func (c *Client) EnsureHostname(clientName string) error {
	if clientName == "" {
		return nil
	}
	clientHost, err := c.GetClientHost(clientName)
	if err != nil || clientHost == "" {
		return nil // no host restriction or can't read spec — proceed anyway
	}
	currentHost, err := os.Hostname()
	if err != nil || clientHost == currentHost {
		return nil
	}
	return c.UpdateClientHost(clientName, currentHost)
}

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
// If client is non-empty it is passed via the top-level -c flag so the
// operation is authorised under that client's context.
func (c *Client) ShelveDelete(cl, client string) error {
	if client != "" {
		_, err := c.exec.Run("-c", client, "shelve", "-d", "-c", cl)
		return err
	}
	_, err := c.exec.Run("shelve", "-d", "-c", cl)
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
// If client is non-empty it is passed via the top-level -c flag.
func (c *Client) Revert(client string, files []string) error {
	if len(files) == 0 {
		return nil
	}
	var args []string
	if client != "" {
		args = append([]string{"-c", client, "revert"}, files...)
	} else {
		args = append([]string{"revert"}, files...)
	}
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
// If client is non-empty it is passed via the top-level -c flag.
func (c *Client) DeleteChange(cl, client string) error {
	if client != "" {
		_, err := c.exec.Run("-c", client, "change", "-d", cl)
		return err
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

// Resolve runs `p4 resolve` attached to the current process's stdio so the
// user can interact with merge prompts. Returns an error if the command exits
// with a non-zero status.
func (c *Client) Resolve() error {
	cmd := exec.Command("p4", "resolve")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("p4 resolve: %s", msg)
	}
	return nil
}

// FindUntracked returns files in dirs that p4 does not know about.
// Uses Go's filepath.WalkDir instead of the Unix `find` command so it
// works on Windows. Files are checked in batches of 50 via p4 fstat.
func (c *Client) FindUntracked(dirs []string, maxDepth int) ([]string, error) {
	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	const batchSize = 50

	// Collect candidate files first.
	var candidates []string
	for _, root := range dirs {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // skip unreadable entries
			}
			if d.IsDir() {
				// Enforce maxDepth when set.
				// Matches find -maxdepth semantics: depth 1 = direct children of root.
				if maxDepth > 0 {
					rel, relErr := filepath.Rel(root, path)
					if relErr == nil && rel != "." {
						// Count path components: "a/b" → depth 2
						depth := len(strings.Split(filepath.ToSlash(rel), "/"))
						if depth > maxDepth {
							return filepath.SkipDir
						}
					}
				}
				return nil
			}
			candidates = append(candidates, path)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk %s: %w", root, err)
		}
	}

	// Batch fstat queries.
	// p4 fstat writes one block per known file to stdout; unknown files
	// produce a "no such file(s)" message on stderr and a non-zero exit.
	// We collect which local paths the server knows about (clientFile /
	// localPath fields), then mark anything missing as untracked.
	known := make(map[string]bool)
	for i := 0; i < len(candidates); i += batchSize {
		end := i + batchSize
		if end > len(candidates) {
			end = len(candidates)
		}
		batch := candidates[i:end]
		args := append([]string{"fstat"}, batch...)
		out, _ := c.exec.Run(args...) // ignore exit error: happens when some files unknown
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			// Match "... clientFile <path>" or "... localPath <path>"
			for _, prefix := range []string{"... clientFile ", "... localPath "} {
				if strings.HasPrefix(line, prefix) {
					p := strings.TrimPrefix(line, prefix)
					known[filepath.ToSlash(p)] = true
					known[filepath.Clean(p)] = true
				}
			}
		}
	}

	var untracked []string
	for _, p := range candidates {
		abs, _ := filepath.Abs(p)
		if !known[filepath.ToSlash(p)] && !known[filepath.Clean(p)] &&
			!known[filepath.ToSlash(abs)] && !known[filepath.Clean(abs)] {
			untracked = append(untracked, p)
		}
	}
	return untracked, nil
}
