package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

// traceCmd is the top-level trace command.
var traceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Monitor and inspect traces",
	Long: `Monitor and inspect execution traces in AegisLab.

A trace represents a top-level workflow (e.g. fault injection pipeline).
Each trace contains child tasks that execute sequentially or in parallel.

WORKFLOW:
  # List recent traces
  aegisctl trace list
  aegisctl trace list --project pair_diagnosis --state Running

  # Get trace details (includes child tasks)
  aegisctl trace get <trace-id>

  # Watch trace events in real-time via SSE
  aegisctl trace watch <trace-id>

TRACE STATES: Pending, Running, Completed, Failed
TRACE TYPES:  FullPipeline, FaultInjection, DatapackBuild, AlgorithmRun`,
}

// --- trace list ---

var traceListProject string
var traceListState string
var traceListGroupID string
var traceListPage int
var traceListSize int

var traceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List traces with optional filters",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		// Resolve project name to ID if given.
		projectIDStr := ""
		projectName := traceListProject
		if projectName == "" {
			projectName = flagProject
		}
		if projectName != "" {
			id, err := newResolver().ProjectID(projectName)
			if err != nil {
				output.PrintInfo(fmt.Sprintf("Warning: could not resolve project %q, using as-is: %v", projectName, err))
				projectIDStr = projectName
			} else {
				projectIDStr = fmt.Sprintf("%d", id)
			}
		}

		path := "/api/v2/traces"
		params := buildQueryParams(map[string]string{
			"project_id": projectIDStr,
			"state":      traceListState,
			"group_id":   traceListGroupID,
			"page":       intToString(traceListPage),
			"size":       intToString(traceListSize),
		})
		if params != "" {
			path += "?" + params
		}

		var resp client.APIResponse[client.PaginatedData[map[string]any]]
		if err := c.Get(path, &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		headers := []string{"TRACE-ID", "TYPE", "STATE", "PROJECT", "START-TIME", "LEAF-NUM"}
		var rows [][]string
		for _, item := range resp.Data.Items {
			rows = append(rows, []string{
				stringField(item, "trace_id"),
				stringField(item, "type"),
				stringField(item, "state"),
				stringField(item, "project_id"),
				stringField(item, "start_time"),
				stringField(item, "leaf_num"),
			})
		}

		output.PrintTable(headers, rows)
		p := resp.Data.Pagination
		output.PrintInfo(fmt.Sprintf("Page %d/%d (total: %d)", p.Page, p.TotalPages, p.Total))
		return nil
	},
}

// --- trace get ---

var traceGetCmd = &cobra.Command{
	Use:   "get <trace-id>",
	Short: "Show detailed trace information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		path := fmt.Sprintf("/api/v2/traces/%s", args[0])
		var resp client.APIResponse[map[string]any]
		if err := c.Get(path, &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		// Print trace header fields.
		for k, v := range resp.Data {
			if k == "tasks" {
				continue
			}
			fmt.Printf("%-20s %v\n", k+":", v)
		}

		// Print child tasks as a table if present.
		if tasksRaw, ok := resp.Data["tasks"]; ok && tasksRaw != nil {
			fmt.Println()
			fmt.Println("Tasks:")

			// Re-marshal and unmarshal to get typed slice.
			data, err := json.Marshal(tasksRaw)
			if err == nil {
				var tasks []map[string]any
				if json.Unmarshal(data, &tasks) == nil && len(tasks) > 0 {
					headers := []string{"TASK-ID", "TYPE", "STATE", "CREATED"}
					var rows [][]string
					for _, t := range tasks {
						rows = append(rows, []string{
							stringField(t, "task_id"),
							stringField(t, "type"),
							stringField(t, "state"),
							stringField(t, "created_at"),
						})
					}
					output.PrintTable(headers, rows)
				}
			}
		}
		return nil
	},
}

// --- trace watch ---

var traceWatchCmd = &cobra.Command{
	Use:   "watch <trace-id>",
	Short: "Watch trace events via SSE stream",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		traceID := args[0]
		ssePath := fmt.Sprintf("/api/v2/traces/%s/stream", traceID)
		reader := client.NewSSEReader(flagServer, ssePath, flagToken)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle Ctrl+C.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		go func() {
			<-sigCh
			cancel()
		}()

		output.PrintInfo(fmt.Sprintf("Watching trace %s (Ctrl+C to stop)...", traceID))

		events, errs := reader.Stream(ctx)

		for {
			select {
			case evt, ok := <-events:
				if !ok {
					return nil
				}
				printSSEEvent(evt)

				// Check for terminal state in event data.
				if isTerminalEvent(evt) {
					output.PrintInfo("Trace reached terminal state.")
					return nil
				}
			case err, ok := <-errs:
				if !ok {
					return nil
				}
				return err
			case <-ctx.Done():
				return nil
			}
		}
	},
}

// printSSEEvent formats and prints an SSE event to stdout.
func printSSEEvent(evt client.SSEEvent) {
	ts := time.Now().Format("15:04:05")

	// Try to parse structured data from the event.
	var data map[string]any
	if json.Unmarshal([]byte(evt.Data), &data) == nil {
		taskType := stringField(data, "task_type")
		taskID := stringField(data, "task_id")
		eventName := evt.Event
		if eventName == "" {
			eventName = stringField(data, "event")
		}

		payload := evt.Data
		if taskType != "" || taskID != "" {
			fmt.Printf("[%s] %-18s %-36s %-20s %s\n", ts, taskType, taskID, eventName, payload)
			return
		}
	}

	// Fallback: print raw event.
	if evt.Event != "" {
		fmt.Printf("[%s] event=%s %s\n", ts, evt.Event, evt.Data)
	} else {
		fmt.Printf("[%s] %s\n", ts, evt.Data)
	}
}

// isTerminalEvent checks if the SSE event indicates the trace has finished.
func isTerminalEvent(evt client.SSEEvent) bool {
	if evt.Event == "completed" || evt.Event == "failed" || evt.Event == "cancelled" {
		return true
	}

	var data map[string]any
	if json.Unmarshal([]byte(evt.Data), &data) == nil {
		state := stringField(data, "state")
		switch state {
		case "Completed", "Failed", "Cancelled", "Error":
			return true
		}
	}
	return false
}

func init() {
	traceListCmd.Flags().StringVar(&traceListProject, "project", "", "Filter by project name or ID")
	traceListCmd.Flags().StringVar(&traceListState, "state", "", "Filter by state (Pending, Running, Completed, Failed)")
	traceListCmd.Flags().StringVar(&traceListGroupID, "group-id", "", "Filter by group ID")
	traceListCmd.Flags().IntVar(&traceListPage, "page", 0, "Page number")
	traceListCmd.Flags().IntVar(&traceListSize, "size", 0, "Page size")

	traceCmd.AddCommand(traceListCmd)
	traceCmd.AddCommand(traceGetCmd)
	traceCmd.AddCommand(traceWatchCmd)
}
