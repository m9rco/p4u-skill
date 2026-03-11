// Package ui provides terminal UI helpers for p4u-skill.
package ui

import (
	"os"

	"github.com/fatih/color"
)

// Colors used throughout the tool.
var (
	Red       = color.New(color.FgRed)
	Green     = color.New(color.FgGreen)
	Yellow    = color.New(color.FgYellow)
	Blue      = color.New(color.FgBlue)
	Cyan      = color.New(color.FgCyan)
	Purple    = color.New(color.FgMagenta)
	White     = color.New(color.FgWhite)
	Bold      = color.New(color.Bold)
	Underline = color.New(color.Underline)
)

// NoColor disables all color output.
func NoColor() {
	color.NoColor = true
}

// ForceColor enables color even when not a TTY (e.g. when piping).
func ForceColor() {
	color.NoColor = false
}

// InitColors sets up colors based on environment.
// Respects NO_COLOR env variable (https://no-color.org/).
func InitColors(noColor, forceColor bool) {
	if noColor || os.Getenv("NO_COLOR") != "" {
		NoColor()
	} else if forceColor {
		ForceColor()
	}
	// Otherwise fatih/color auto-detects terminal.
}

// Sprint returns a colored string using the given color attributes.
func Sprint(c *color.Color, s string) string {
	return c.Sprint(s)
}

// Sprintf returns a colored formatted string.
func Sprintf(c *color.Color, format string, a ...interface{}) string {
	return c.Sprintf(format, a...)
}
