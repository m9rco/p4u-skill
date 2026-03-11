package p4

import (
	"strings"
)

// Info holds the parsed output of `p4 info`.
type Info struct {
	UserName   string
	ClientName string
	ClientRoot string
	HostName   string
}

// GetInfo parses `p4 info` and returns structured data.
func (c *Client) GetInfo() (*Info, error) {
	out, err := c.exec.Run("info")
	if err != nil {
		return nil, err
	}
	info := &Info{}
	for _, line := range strings.Split(out, "\n") {
		if v, ok := trimPrefix(line, "User name: "); ok {
			info.UserName = v
		} else if v, ok := trimPrefix(line, "Client name: "); ok {
			info.ClientName = v
		} else if v, ok := trimPrefix(line, "Client root: "); ok {
			info.ClientRoot = v
		} else if v, ok := trimPrefix(line, "Client host: "); ok {
			info.HostName = v
		}
	}
	return info, nil
}

// trimPrefix removes the given prefix from s and returns (remainder, true) if
// found, or ("", false) otherwise.
func trimPrefix(s, prefix string) (string, bool) {
	if strings.HasPrefix(s, prefix) {
		return strings.TrimSpace(s[len(prefix):]), true
	}
	return "", false
}
