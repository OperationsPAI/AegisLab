package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// Quiet suppresses informational messages when true.
var Quiet bool

// OutputFormat represents the output format type.
type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
)

// PrintJSON writes v as indented JSON to stdout.
func PrintJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		PrintError(err.Error())
		return
	}
	fmt.Fprintln(os.Stdout, string(data))
}

// PrintTable renders a simple ASCII table to stdout.
func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	_ = w.Flush()
}

// PrintInfo writes an informational message to stderr (suppressed when Quiet).
func PrintInfo(msg string) {
	if !Quiet {
		fmt.Fprintln(os.Stderr, msg)
	}
}

// PrintError writes an error message to stderr.
func PrintError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
}
