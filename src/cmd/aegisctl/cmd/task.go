package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

// taskCmd is the top-level task command.
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Monitor and inspect tasks",
	Long: `Monitor and inspect individual tasks in AegisLab.

Tasks are the atomic units of work executed by the consumer (background worker).
They are typically created as children of a trace.

WORKFLOW:
  # List tasks with filters
  aegisctl task list
  aegisctl task list --state Running --type FaultInjection

  # Get task details
  aegisctl task get <task-id>

  # Stream task logs via WebSocket
  aegisctl task logs <task-id> --follow

TASK STATES: Pending, Rescheduled, Running, Completed, Error, Cancelled
TASK TYPES:  BuildContainer, RestartPedestal, FaultInjection, RunAlgorithm,
             BuildDatapack, CollectResult, CronJob`,
}

// --- task list ---

var taskListState string
var taskListType string
var taskListPage int
var taskListSize int
var taskListOverdue bool

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks with optional filters",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		path := "/api/v2/tasks"
		params := buildQueryParams(map[string]string{
			"state": taskListState,
			"type":  taskListType,
			"page":  intToString(taskListPage),
			"size":  intToString(taskListSize),
		})
		if params != "" {
			path += "?" + params
		}

		var resp client.APIResponse[client.PaginatedData[map[string]any]]
		if err := c.Get(path, &resp); err != nil {
			return err
		}

		// Apply --overdue filter client-side: only keep Pending tasks whose
		// execute_time is in the past. This avoids a new backend filter and
		// works even when the server doesn't know about WAIT semantics.
		items := resp.Data.Items
		if taskListOverdue {
			nowEpoch := time.Now().Unix()
			filtered := items[:0]
			for _, item := range items {
				if stringField(item, "state") != "Pending" {
					continue
				}
				if execTimeField(item) <= nowEpoch {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(items)
			return nil
		}

		headers := []string{"TASK-ID", "TYPE", "STATE", "WAIT", "TRACE-ID", "PROJECT-ID", "CREATED"}
		var rows [][]string
		nowEpoch := time.Now().Unix()
		for _, item := range items {
			state := stringField(item, "state")
			wait := "-"
			if state == "Pending" {
				wait = formatWait(execTimeField(item) - nowEpoch)
			}
			rows = append(rows, []string{
				stringField(item, "id"),
				stringField(item, "type"),
				state,
				wait,
				stringField(item, "trace_id"),
				stringField(item, "project_id"),
				stringField(item, "created_at"),
			})
		}

		output.PrintTable(headers, rows)
		p := resp.Data.Pagination
		output.PrintInfo(fmt.Sprintf("Page %d/%d (total: %d)", p.Page, p.TotalPages, p.Total))
		return nil
	},
}

// --- task expedite ---

var taskExpediteCmd = &cobra.Command{
	Use:   "expedite <task-id>",
	Short: "Expedite a Pending task so it runs on the next scheduler tick",
	Long: `Expedite a Pending task.

Resets execute_time to now in both MySQL and the Redis delayed queue, so the
scheduler picks the task up on its next tick. Rejects the call if the task is
in any state other than Pending. Idempotent: expediting an already-due task
succeeds silently.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		path := fmt.Sprintf("/api/v2/tasks/%s/expedite", args[0])

		var resp client.APIResponse[map[string]any]
		if err := c.Post(path, nil, &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		output.PrintInfo(fmt.Sprintf("Task %s expedited", args[0]))
		return nil
	},
}

// formatWait renders a remaining-time int64 second count as "+MM:SS" or
// "-MM:SS" (overdue). Used only for Pending rows.
func formatWait(deltaSec int64) string {
	sign := "+"
	abs := deltaSec
	if deltaSec < 0 {
		sign = "-"
		abs = -deltaSec
	}
	return fmt.Sprintf("%s%02d:%02d", sign, abs/60, abs%60)
}

// execTimeField extracts execute_time from a task response map. JSON numbers
// decode as float64 through encoding/json, so handle that plus a couple of
// fallback numeric types defensively.
func execTimeField(m map[string]any) int64 {
	v, ok := m["execute_time"]
	if !ok || v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case int:
		return int64(n)
	}
	return 0
}

// --- task get ---

var taskGetCmd = &cobra.Command{
	Use:   "get <task-id>",
	Short: "Show detailed task information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		path := fmt.Sprintf("/api/v2/tasks/%s", args[0])
		var resp client.APIResponse[map[string]any]
		if err := c.Get(path, &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		// Print all fields as key-value pairs.
		for k, v := range resp.Data {
			fmt.Printf("%-20s %v\n", k+":", v)
		}
		return nil
	},
}

// --- task logs ---

var taskLogsFollow bool

var taskLogsCmd = &cobra.Command{
	Use:   "logs <task-id>",
	Short: "Stream task logs via WebSocket",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		wsPath := fmt.Sprintf("/api/v2/tasks/%s/logs/ws", taskID)
		reader := client.NewWSReader(flagServer, wsPath, flagToken)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle Ctrl+C.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		go func() {
			<-sigCh
			cancel()
		}()

		messages, errs := reader.Stream(ctx)

		if taskLogsFollow {
			// Follow mode: keep reading until cancelled.
			for {
				select {
				case msg, ok := <-messages:
					if !ok {
						return nil
					}
					fmt.Println(msg)
				case err, ok := <-errs:
					if !ok {
						return nil
					}
					return err
				case <-ctx.Done():
					return nil
				}
			}
		} else {
			// Non-follow mode: read available messages with a timeout.
			timeout := time.After(5 * time.Second)
			for {
				select {
				case msg, ok := <-messages:
					if !ok {
						return nil
					}
					fmt.Println(msg)
				case err, ok := <-errs:
					if !ok {
						return nil
					}
					return err
				case <-timeout:
					return nil
				case <-ctx.Done():
					return nil
				}
			}
		}
	},
}

// Helper functions shared by task and trace commands.

func buildQueryParams(params map[string]string) string {
	var parts []string
	for k, v := range params {
		if v != "" && v != "0" {
			parts = append(parts, k+"="+v)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += "&" + p
	}
	return result
}

func intToString(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d", n)
}

func stringField(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func init() {
	taskListCmd.Flags().StringVar(&taskListState, "state", "", "Filter by state (Pending, Running, Completed, Error, Cancelled, Rescheduled)")
	taskListCmd.Flags().StringVar(&taskListType, "type", "", "Filter by type (BuildContainer, RestartPedestal, FaultInjection, RunAlgorithm, BuildDatapack, CollectResult, CronJob)")
	taskListCmd.Flags().IntVar(&taskListPage, "page", 0, "Page number")
	taskListCmd.Flags().IntVar(&taskListSize, "size", 0, "Page size")
	taskListCmd.Flags().BoolVar(&taskListOverdue, "overdue", false, "Show only Pending tasks whose execute_time has passed (WAIT < 0)")

	taskLogsCmd.Flags().BoolVarP(&taskLogsFollow, "follow", "f", false, "Follow log output")

	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskGetCmd)
	taskCmd.AddCommand(taskLogsCmd)
	taskCmd.AddCommand(taskExpediteCmd)
}
