package cmd

import (
	"fmt"
	"os"
	"strings"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

// rateLimiterCmd groups `rate-limiter` subcommands (OperationsPAI/aegis#21).
var rateLimiterCmd = &cobra.Command{
	Use:     "rate-limiter",
	Aliases: []string{"rl"},
	Short:   "Inspect and manage token-bucket rate limiters",
	Long: `Inspect and manage the token-bucket rate limiters used by the consumer.

Each bucket is a Redis set keyed "token_bucket:<name>" whose members are
the task_ids currently holding a token. If a task crashes without
releasing its token, the bucket can get stuck at HELD=CAP and block all
later tasks.

COMMANDS:
  aegisctl rate-limiter status
  aegisctl rate-limiter reset --bucket NAME --force
  aegisctl rate-limiter gc`,
}

type rlHolder struct {
	TaskID     string `json:"task_id"`
	TaskState  string `json:"task_state"`
	IsTerminal bool   `json:"is_terminal"`
}

type rlItem struct {
	Bucket   string     `json:"bucket"`
	Key      string     `json:"key"`
	Capacity int        `json:"capacity"`
	Held     int        `json:"held"`
	Holders  []rlHolder `json:"holders"`
}

type rlListResp struct {
	Items []rlItem `json:"items"`
}

type rlGCResp struct {
	Released       int `json:"released"`
	TouchedBuckets int `json:"touched_buckets"`
}

const (
	ansiRed   = "\x1b[31m"
	ansiReset = "\x1b[0m"
)

func useColor() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if fi, err := os.Stdout.Stat(); err == nil {
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

var rateLimiterStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "List token-bucket rate limiters, their holders, and DB state",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		var resp client.APIResponse[rlListResp]
		if err := c.Get("/api/v2/rate-limiters", &resp); err != nil {
			return err
		}
		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}
		color := useColor()
		headers := []string{"BUCKET", "HELD/CAP", "HOLDERS"}
		var rows [][]string
		for _, item := range resp.Data.Items {
			var parts []string
			for _, h := range item.Holders {
				s := fmt.Sprintf("%s[%s]", h.TaskID, h.TaskState)
				if h.IsTerminal {
					if color {
						s = ansiRed + s + " (LEAKED)" + ansiReset
					} else {
						s = s + " (LEAKED)"
					}
				}
				parts = append(parts, s)
			}
			holders := strings.Join(parts, ", ")
			if holders == "" {
				holders = "-"
			}
			rows = append(rows, []string{
				item.Bucket, fmt.Sprintf("%d/%d", item.Held, item.Capacity), holders,
			})
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

var (
	rlResetBucket string
	rlResetForce  bool
)

var rateLimiterResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Delete a rate-limiter bucket key (requires --force)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if rlResetBucket == "" {
			return fmt.Errorf("--bucket is required")
		}
		if !rlResetForce {
			return fmt.Errorf("refusing to reset bucket %q without --force", rlResetBucket)
		}
		c := newClient()
		var resp client.APIResponse[any]
		if err := c.Delete("/api/v2/rate-limiters/"+rlResetBucket, &resp); err != nil {
			return err
		}
		output.PrintInfo(fmt.Sprintf("bucket %q reset", rlResetBucket))
		return nil
	},
}

var rateLimiterGCCmd = &cobra.Command{
	Use:   "gc",
	Short: "Release tokens held by terminal-state tasks across all buckets",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		var resp client.APIResponse[rlGCResp]
		if err := c.Post("/api/v2/rate-limiters/gc", nil, &resp); err != nil {
			return err
		}
		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}
		output.PrintInfo(fmt.Sprintf("released %d leaked tokens from %d buckets",
			resp.Data.Released, resp.Data.TouchedBuckets))
		return nil
	},
}

func init() {
	rateLimiterResetCmd.Flags().StringVar(&rlResetBucket, "bucket", "", "Bucket short name, e.g. restart_service")
	rateLimiterResetCmd.Flags().BoolVar(&rlResetForce, "force", false, "Required to actually perform the reset")
	rateLimiterCmd.AddCommand(rateLimiterStatusCmd)
	rateLimiterCmd.AddCommand(rateLimiterResetCmd)
	rateLimiterCmd.AddCommand(rateLimiterGCCmd)
	rootCmd.AddCommand(rateLimiterCmd)
}
