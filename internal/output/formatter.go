// Package output provides formatting for p4u command output.
package output

import (
	"encoding/json"
	"fmt"
	"os"
)

// Format is the output format type.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Printer writes structured results to stdout.
type Printer struct {
	format Format
}

// New creates a Printer.
func New(jsonMode bool) *Printer {
	if jsonMode {
		return &Printer{format: FormatJSON}
	}
	return &Printer{format: FormatText}
}

// IsJSON returns true when in JSON mode.
func (p *Printer) IsJSON() bool { return p.format == FormatJSON }

// PrintJSON marshals v to JSON and writes it to stdout.
func (p *Printer) PrintJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// PrintError writes an error message. In JSON mode emits {"error":"..."}.
func (p *Printer) PrintError(err error) {
	if p.IsJSON() {
		p.PrintJSON(map[string]string{"error": err.Error()})
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
}

// PrintText writes plain text to stdout.
func (p *Printer) PrintText(s string) {
	fmt.Print(s)
}

// PrintTextLn writes plain text with newline to stdout.
func (p *Printer) PrintTextLn(s string) {
	fmt.Println(s)
}
