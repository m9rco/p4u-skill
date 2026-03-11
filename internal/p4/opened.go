package p4

import (
	"fmt"
	"strings"
)

// OpenedFile represents a file opened in a changelist.
type OpenedFile struct {
	DepotPath  string
	Changelist string // "default" or number
	Action     string
}

// OpenedFiles returns all files opened by a user/client.
func (c *Client) OpenedFiles(client string) ([]OpenedFile, error) {
	args := []string{"opened"}
	if client != "" {
		args = append(args, "-C", client)
	}
	out, err := c.exec.Run(args...)
	if err != nil {
		return nil, nil // no opened files is OK
	}
	var files []OpenedFile
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		f := parseOpenedLine(line)
		files = append(files, f)
	}
	return files, nil
}

// OpenedInChangelist returns the depot paths of files opened in a specific changelist.
func (c *Client) OpenedInChangelist(cl, client string) ([]string, error) {
	files, err := c.OpenedFiles(client)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, f := range files {
		if f.Changelist == cl {
			paths = append(paths, f.DepotPath)
		}
	}
	return paths, nil
}

// GetChangeClient returns the client name associated with a changelist.
func (c *Client) GetChangeClient(cl string) (string, error) {
	out, err := c.exec.Run("change", "-o", cl)
	if err != nil {
		return "", fmt.Errorf("change -o %s: %w", cl, err)
	}
	for _, line := range strings.Split(out, "\n") {
		if v, ok := trimPrefix(line, "Client:"); ok {
			return v, nil
		}
	}
	return "", fmt.Errorf("could not find Client field in change %s", cl)
}

// GetClientHost returns the host field of a client spec.
func (c *Client) GetClientHost(clientName string) (string, error) {
	out, err := c.exec.Run("client", "-o", clientName)
	if err != nil {
		return "", fmt.Errorf("client -o %s: %w", clientName, err)
	}
	for _, line := range strings.Split(out, "\n") {
		if v, ok := trimPrefix(line, "Host:"); ok {
			return v, nil
		}
	}
	return "", nil // Host field may be empty
}

// FixHostname updates the hostname field of a client spec to the current host.
func (c *Client) FixHostname(clientName, newHost string) error {
	out, err := c.exec.Run("client", "-o", clientName)
	if err != nil {
		return err
	}
	var updated strings.Builder
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Host:") {
			updated.WriteString(fmt.Sprintf("Host:\t%s\n", newHost))
		} else {
			updated.WriteString(line + "\n")
		}
	}
	_, err = c.exec.Run("client", "-i")
	_ = updated
	// Use direct exec for stdin-based command.
	return runWithStdin(clientName, "client", "-i", updated.String())
}

// parseOpenedLine parses a line from `p4 opened`.
// Format: "//depot/path#rev - action change NNNN (filetype)"
func parseOpenedLine(line string) OpenedFile {
	f := OpenedFile{}
	// Extract depot path (up to #).
	if idx := strings.Index(line, "#"); idx >= 0 {
		f.DepotPath = strings.TrimSpace(line[:idx])
		rest := line[idx:]
		// rest: "#rev - action change NNNN ..."
		parts := strings.Fields(rest)
		// parts[0]="#rev", parts[1]="-", parts[2]=action, parts[3]="change", parts[4]=NNNN or "default"
		if len(parts) >= 5 {
			f.Action = parts[2]
			if parts[3] == "default" {
				f.Changelist = "default"
			} else if parts[3] == "change" && len(parts) >= 5 {
				f.Changelist = parts[4]
			}
		}
	}
	return f
}

// GetFixes returns job/fix IDs associated with a changelist.
func (c *Client) GetFixes(cl string) ([]string, error) {
	out, err := c.exec.Run("fixes", "-c", cl)
	if err != nil {
		return nil, nil
	}
	var fixes []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 0 {
			fixes = append(fixes, parts[0])
		}
	}
	return fixes, nil
}

// ListClients returns the client names for a given user.
func (c *Client) ListClients(user string) ([]string, error) {
	args := []string{"clients", "-u", user}
	out, err := c.exec.Run(args...)
	if err != nil {
		return nil, err
	}
	var clients []string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Format: "Client name date root description"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			clients = append(clients, parts[1])
		}
	}
	return clients, nil
}

// GetClientPath returns the root path of a client spec.
func (c *Client) GetClientPath(clientName string) (string, error) {
	out, err := c.exec.Run("client", "-o", clientName)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(out, "\n") {
		if v, ok := trimPrefix(line, "Root:"); ok {
			return v, nil
		}
	}
	return "", nil
}
