package ui

import (
	"fmt"
	"os"
	"strings"
)

const progressWidth = 20

// Progress renders a simple text progress bar to stderr.
type Progress struct {
	total   int
	current int
	label   string
}

// NewProgress creates a progress tracker with the given total steps.
func NewProgress(total int, label string) *Progress {
	return &Progress{total: total, label: label}
}

// Advance increments the progress by n steps and redraws.
func (p *Progress) Advance(n int) {
	p.current += n
	p.draw()
}

// Clear removes the progress bar line.
func (p *Progress) Clear() {
	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", progressWidth+len(p.label)+20))
}

func (p *Progress) draw() {
	if p.total <= 0 {
		return
	}
	filled := progressWidth * p.current / p.total
	if filled > progressWidth {
		filled = progressWidth
	}
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", progressWidth-filled)
	pct := 100 * p.current / p.total
	fmt.Fprintf(os.Stderr, "\r%s[%s] (%d%%)", Cyan.Sprint(p.label+" "), bar, pct)
}
