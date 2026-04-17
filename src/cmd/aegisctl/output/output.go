package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

var Quiet bool

type OutputFormat string

const (
	FormatJSON  OutputFormat = "json"
	FormatTable OutputFormat = "table"
)

func PrintInfo(msg string) {
	if Quiet {
		return
	}
	fmt.Fprintln(os.Stderr, msg)
}

func PrintError(err error) {
	fmt.Fprintln(os.Stderr, "Error: "+err.Error())
}

func PrintJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		PrintError(err)
		return
	}
	fmt.Fprintln(os.Stdout, string(data))
}

func PrintTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}
