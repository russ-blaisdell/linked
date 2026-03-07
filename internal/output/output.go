// Package output renders linked results in pretty, JSON, or table format.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// Format is the output rendering mode.
type Format string

const (
	FormatPretty Format = "pretty"
	FormatJSON   Format = "json"
	FormatTable  Format = "table"
)

// ParseFormat validates and returns a Format.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "pretty", "":
		return FormatPretty, nil
	case "json":
		return FormatJSON, nil
	case "table":
		return FormatTable, nil
	default:
		return "", fmt.Errorf("unknown output format %q — use pretty, json, or table", s)
	}
}

// Printer renders output to a writer (defaults to os.Stdout).
type Printer struct {
	w      io.Writer
	format Format
}

// New returns a Printer for the given format writing to stdout.
func New(format Format) *Printer {
	return &Printer{w: os.Stdout, format: format}
}

// NewWithWriter returns a Printer writing to w.
func NewWithWriter(format Format, w io.Writer) *Printer {
	return &Printer{w: w, format: format}
}

// JSON encodes v as indented JSON.
func (p *Printer) JSON(v interface{}) error {
	enc := json.NewEncoder(p.w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// Table renders rows with a header row.
func (p *Printer) Table(headers []string, rows [][]string) {
	tw := tablewriter.NewWriter(p.w)
	tw.SetHeader(headers)
	tw.SetBorder(false)
	tw.SetColumnSeparator("  ")
	tw.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tw.SetAlignment(tablewriter.ALIGN_LEFT)
	tw.SetHeaderLine(true)
	tw.AppendBulk(rows)
	tw.Render()
}

// Println writes a plain line.
func (p *Printer) Println(a ...interface{}) {
	fmt.Fprintln(p.w, a...)
}

// Printf writes a formatted line.
func (p *Printer) Printf(format string, a ...interface{}) {
	fmt.Fprintf(p.w, format, a...)
}

// Header prints a bold section header (pretty mode only).
func (p *Printer) Header(s string) {
	if p.format == FormatJSON {
		return
	}
	bold := color.New(color.Bold)
	bold.Fprintln(p.w, s)
}

// Field prints a labeled field in pretty mode.
func (p *Printer) Field(label, value string) {
	if value == "" {
		return
	}
	labelColor := color.New(color.FgCyan, color.Bold)
	labelColor.Fprintf(p.w, "  %-18s", label+":")
	fmt.Fprintln(p.w, value)
}

// Success prints a green success message.
func (p *Printer) Success(msg string) {
	if p.format == FormatJSON {
		_ = p.JSON(map[string]string{"status": "ok", "message": msg})
		return
	}
	color.New(color.FgGreen).Fprintln(p.w, "✓ "+msg)
}

// Error prints a red error message to stderr.
func Error(msg string) {
	color.New(color.FgRed).Fprintln(os.Stderr, "✗ "+msg)
}

// Warn prints a yellow warning.
func (p *Printer) Warn(msg string) {
	color.New(color.FgYellow).Fprintln(p.w, "⚠ "+msg)
}

// Format returns the configured output format.
func (p *Printer) Format() Format {
	return p.format
}
