package cmd

import (
	"fmt"
	"os"
	"time"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

var (
	waitTimeout  int
	waitInterval int
)

var waitCmd = &cobra.Command{
	Use:   "wait <trace-id|task-id>",
	Short: "Block until a trace or task reaches a terminal state",
	Long: `Block until a trace or task reaches a terminal state.

Automatically detects whether the ID refers to a trace or task.
Polls the API at regular intervals and prints status updates to stderr.

EXIT CODES:
  0 — Completed successfully
  2 — Failed, Error, or Cancelled
  3 — Timeout (exceeded --timeout)

EXAMPLES:
  # Wait for a trace to complete (default timeout: 600s)
  aegisctl wait <trace-id>

  # Wait with custom timeout and poll interval
  aegisctl wait <trace-id> --timeout 1200 --interval 10

  # Use in scripts to chain operations
  aegisctl inject submit --spec inject.yaml --project my_project -o json | \
    jq -r '.trace_id' | xargs aegisctl wait`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		c := newClient()

		// Determine resource type: try trace first, then task.
		resourceType, err := detectResourceType(c, id)
		if err != nil {
			return err
		}

		deadline := time.Now().Add(time.Duration(waitTimeout) * time.Second)
		interval := time.Duration(waitInterval) * time.Second
		start := time.Now()

		for {
			state, data, err := pollState(c, resourceType, id)
			if err != nil {
				return fmt.Errorf("poll %s %s: %w", resourceType, id, err)
			}

			elapsed := time.Since(start).Truncate(time.Second)
			output.PrintInfo(fmt.Sprintf("Waiting for %s... [%s] %s", id, state, elapsed))

			if isTerminal(resourceType, state) {
				// Print final result.
				if output.OutputFormat(flagOutput) == output.FormatJSON {
					output.PrintJSON(data)
				} else {
					output.PrintTable(
						[]string{"ID", "TYPE", "STATE", "ELAPSED"},
						[][]string{{id, resourceType, state, elapsed.String()}},
					)
				}

				// Determine exit code.
				switch state {
				case "Completed":
					return nil
				case "Failed", "Error":
					os.Exit(2)
				case "Cancelled":
					os.Exit(2)
				}
				return nil
			}

			if time.Now().After(deadline) {
				output.PrintError(fmt.Errorf("timeout after %ds waiting for %s %s (last state: %s)", waitTimeout, resourceType, id, state))
				os.Exit(3)
			}

			time.Sleep(interval)
		}
	},
}

// detectResourceType probes the API to determine if the id refers to a trace
// or a task. Returns "trace" or "task".
func detectResourceType(c *client.Client, id string) (string, error) {
	var traceResp client.APIResponse[any]
	err := c.Get(fmt.Sprintf("/api/v2/traces/%s", id), &traceResp)
	if err == nil {
		return "trace", nil
	}

	// Check if it was a 404.
	if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
		var taskResp client.APIResponse[any]
		err2 := c.Get(fmt.Sprintf("/api/v2/tasks/%s", id), &taskResp)
		if err2 == nil {
			return "task", nil
		}
		if apiErr2, ok2 := err2.(*client.APIError); ok2 && apiErr2.StatusCode == 404 {
			return "", fmt.Errorf("neither trace nor task found for id %q", id)
		}
		return "", fmt.Errorf("lookup task %s: %w", id, err2)
	}

	return "", fmt.Errorf("lookup trace %s: %w", id, err)
}

// stateResponse is a minimal struct to extract the state field from API responses.
type stateResponse struct {
	State string `json:"state"`
}

// pollState fetches the current state and full data for the given resource.
func pollState(c *client.Client, resourceType, id string) (string, any, error) {
	var path string
	switch resourceType {
	case "trace":
		path = fmt.Sprintf("/api/v2/traces/%s", id)
	case "task":
		path = fmt.Sprintf("/api/v2/tasks/%s", id)
	default:
		return "", nil, fmt.Errorf("unknown resource type: %s", resourceType)
	}

	var resp client.APIResponse[map[string]any]
	if err := c.Get(path, &resp); err != nil {
		return "", nil, err
	}

	state, _ := resp.Data["state"].(string)
	return state, resp.Data, nil
}

// isTerminal returns true if the state is a terminal state for the given
// resource type.
func isTerminal(resourceType, state string) bool {
	switch resourceType {
	case "trace":
		return state == "Completed" || state == "Failed"
	case "task":
		return state == "Completed" || state == "Error" || state == "Cancelled"
	}
	return false
}

func init() {
	waitCmd.Flags().IntVar(&waitTimeout, "timeout", 600, "Maximum time to wait in seconds")
	waitCmd.Flags().IntVar(&waitInterval, "interval", 5, "Poll interval in seconds")
}
