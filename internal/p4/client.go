// Package p4 provides a wrapper around the p4 CLI tool.
package p4

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Executor runs p4 commands.
type Executor interface {
	Run(args ...string) (string, error)
}

// CLIExecutor runs real p4 commands.
type CLIExecutor struct{}

// Run executes a p4 command and returns its stdout output.
func (e *CLIExecutor) Run(args ...string) (string, error) {
	cmd := exec.Command("p4", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("p4 %s: %s", strings.Join(args, " "), msg)
	}
	return strings.TrimRight(stdout.String(), "\n"), nil
}

// Client wraps a p4 executor for high-level operations.
type Client struct {
	exec Executor
}

// New creates a new Client using the real p4 CLI.
func New() *Client {
	return &Client{exec: &CLIExecutor{}}
}

// NewWithExecutor creates a Client with a custom executor (useful for testing).
func NewWithExecutor(e Executor) *Client {
	return &Client{exec: e}
}

// Run executes an arbitrary p4 command and returns stdout.
func (c *Client) Run(args ...string) (string, error) {
	return c.exec.Run(args...)
}
