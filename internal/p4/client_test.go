package p4_test

import (
	"testing"

	"github.com/m9rco/p4u-skill/internal/p4"
)

// MockExecutor implements p4.Executor for testing.
type MockExecutor struct {
	calls    [][]string
	response map[string]string
	errors   map[string]error
}

func NewMock() *MockExecutor {
	return &MockExecutor{
		response: make(map[string]string),
		errors:   make(map[string]error),
	}
}

func (m *MockExecutor) Set(key, value string) { m.response[key] = value }
func (m *MockExecutor) SetErr(key string, err error) { m.errors[key] = err }

func (m *MockExecutor) Run(args ...string) (string, error) {
	key := args[0]
	if len(args) > 1 {
		key = args[0] + "_" + args[1]
	}
	m.calls = append(m.calls, args)
	if err, ok := m.errors[key]; ok {
		return "", err
	}
	if v, ok := m.response[key]; ok {
		return v, nil
	}
	return "", nil
}

func TestGetInfo(t *testing.T) {
	mock := NewMock()
	mock.Set("info", `
User name: jdoe
Client name: jdoe-ws
Client root: /home/jdoe/ws
Client host: devbox
`)
	c := p4.NewWithExecutor(mock)
	info, err := c.GetInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.UserName != "jdoe" {
		t.Errorf("UserName = %q, want %q", info.UserName, "jdoe")
	}
	if info.ClientName != "jdoe-ws" {
		t.Errorf("ClientName = %q, want %q", info.ClientName, "jdoe-ws")
	}
	if info.ClientRoot != "/home/jdoe/ws" {
		t.Errorf("ClientRoot = %q, want %q", info.ClientRoot, "/home/jdoe/ws")
	}
}

func TestListChanges(t *testing.T) {
	mock := NewMock()
	// With Status=StatusPending, args become ["changes", "-s", "pending"], key="changes_-s"
	mock.Set("changes_-s", `Change 12345 on 2024/01/01 by jdoe@jdoe-ws 'fix bug'
Change 12346 on 2024/01/02 by jdoe@jdoe-ws 'add feature'
`)
	c := p4.NewWithExecutor(mock)
	nums, err := c.ListChanges(p4.ListChangesOpts{Status: p4.StatusPending})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nums) != 2 {
		t.Fatalf("expected 2 changelists, got %d", len(nums))
	}
	if nums[0] != "12345" {
		t.Errorf("nums[0] = %q, want %q", nums[0], "12345")
	}
	if nums[1] != "12346" {
		t.Errorf("nums[1] = %q, want %q", nums[1], "12346")
	}
}

func TestParseDescribe(t *testing.T) {
	mock := NewMock()
	// args: ["describe", "-s", "12345"], key = "describe_-s"
	mock.Set("describe_-s", `Change 12345 by jdoe@jdoe-ws on 2024/01/01 *pending*

	Fix the critical bug in login flow
	ReviewBoard: https://rb.example.com/r/999/

Jobs fixed:
	JOB-001 on 2024/01/01 by jdoe *closed* 'Fix login'

Affected files ...

... //depot/main/login.cpp#5 edit
... //depot/main/auth.h#3 edit
`)
	c := p4.NewWithExecutor(mock)
	d, err := c.Describe("12345", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.IsPending {
		t.Error("expected IsPending=true")
	}
	if len(d.ReviewLinks) == 0 {
		t.Error("expected ReviewLinks to be non-empty")
	}
}
