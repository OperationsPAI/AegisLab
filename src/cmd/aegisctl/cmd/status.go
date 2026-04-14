package cmd

import (
	"fmt"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

// Local structs for status command responses.

type profileInfo struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type taskItem struct {
	ID    int    `json:"id"`
	State string `json:"state"`
}

type traceItem struct {
	TraceID   string `json:"trace_id"`
	State     string `json:"state"`
	Type      string `json:"type"`
	Project   string `json:"project_name"`
	ProjectID int    `json:"project_id"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cluster and connection status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		// Determine context name.
		ctxName := "unknown"
		if cfg != nil && cfg.CurrentContext != "" {
			ctxName = cfg.CurrentContext
		}

		// --- User info ---
		connected := "yes"
		username := "unknown"
		var profileResp client.APIResponse[profileInfo]
		if err := c.Get("/api/v2/auth/profile", &profileResp); err != nil {
			connected = "no (not logged in)"
		} else {
			username = profileResp.Data.Username
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			result := map[string]any{
				"server":    flagServer,
				"context":   ctxName,
				"connected": connected == "yes",
				"username":  username,
			}

			// Tasks
			var taskResp client.APIResponse[client.PaginatedData[taskItem]]
			if err := c.Get("/api/v2/tasks?page=1&size=100", &taskResp); err == nil {
				counts := map[string]int{}
				for _, t := range taskResp.Data.Items {
					counts[t.State]++
				}
				result["tasks"] = counts
			}

			// Traces
			var traceResp client.APIResponse[client.PaginatedData[traceItem]]
			if err := c.Get("/api/v2/traces?page=1&size=10", &traceResp); err == nil {
				result["recent_traces"] = traceResp.Data.Items
			}

			output.PrintJSON(result)
			return nil
		}

		// --- Table output ---
		fmt.Printf("Server:    %s (%s)\n", flagServer, ctxName)
		fmt.Printf("User:      %s\n", username)
		fmt.Printf("Connected: %s\n", connected)
		fmt.Println()

		// --- Tasks ---
		var taskResp client.APIResponse[client.PaginatedData[taskItem]]
		if err := c.Get("/api/v2/tasks?page=1&size=100", &taskResp); err == nil {
			counts := map[string]int{}
			total := 0
			for _, t := range taskResp.Data.Items {
				counts[t.State]++
				total++
			}
			fmt.Printf("Active Tasks:     %d\n", total)
			for state, count := range counts {
				fmt.Printf("  %-14s  %d\n", state+":", count)
			}
		}
		fmt.Println()

		// --- Recent Traces ---
		var traceResp client.APIResponse[client.PaginatedData[traceItem]]
		if err := c.Get("/api/v2/traces?page=1&size=10", &traceResp); err == nil && len(traceResp.Data.Items) > 0 {
			fmt.Println("Recent Traces:")

			rows := make([][]string, 0, len(traceResp.Data.Items))
			for _, t := range traceResp.Data.Items {
				project := t.Project
				if project == "" {
					project = fmt.Sprintf("%d", t.ProjectID)
				}
				rows = append(rows, []string{t.TraceID, t.State, t.Type, project})
			}
			output.PrintTable([]string{"Trace-ID", "State", "Type", "Project"}, rows)
		}

		return nil
	},
}
