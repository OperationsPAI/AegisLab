package output

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
)

// Quiet suppresses informational messages when true.
var Quiet bool

// Format represents an output format.
type Format string

const FormatJSON Format = "json"

// OutputFormat normalizes a string flag to a Format value.
func OutputFormat(s string) Format { return Format(s) }

// PrintJSON marshals v to indented JSON and writes to stdout.
func PrintJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		PrintError(err)
		return
	}
	fmt.Fprintln(os.Stdout, string(data))
}

// PrintTable writes a formatted table to stdout with the given headers and rows.
func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for i, h := range headers {
		if i > 0 {
			fmt.Fprint(w, "\t")
		}
		fmt.Fprint(w, h)
	}
	fmt.Fprintln(w)
	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, col)
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}

// PrintInfo writes an informational message to stderr (suppressed if Quiet).
func PrintInfo(msg string) {
	if Quiet {
		return
	}
	fmt.Fprintf(os.Stderr, "INFO: %s\n", msg)
}

// PrintError writes an error message to stderr.
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
}
