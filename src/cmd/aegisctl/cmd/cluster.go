package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"aegis/cmd/aegisctl/cluster"

	"github.com/spf13/cobra"
)

var (
	clusterPreflightCheck   string
	clusterPreflightFix     bool
	clusterPreflightConfig  string
	clusterPreflightTimeout int
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Inspect and repair AegisLab cluster dependencies",
	Long: `The "cluster" subcommand groups operations that target the AegisLab
cluster and its backing services (Kubernetes, MySQL, ClickHouse, Redis, etcd).`,
}

var clusterPreflightCmd = &cobra.Command{
	Use:   "preflight",
	Short: "Verify that every dependency required by AegisLab is reachable and configured",
	Long: `Runs a catalog of checks against the cluster and the services referenced
by config.dev.toml. The command prints one row per check with status
[OK] / [FAIL] / [WARN] and a suggested fix on failure.

Use --check <id> to run a single check, or --fix to apply idempotent
remediation for the subset of checks that support it (currently:
k8s.rcabench-sa and redis.token-bucket-leaks).

Exit code is 0 when every executed check is OK, 1 otherwise.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cluster.LoadConfig(clusterPreflightConfig)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		env := cluster.NewLiveEnv(cfg)
		reg := cluster.NewRegistry(cluster.DefaultChecks())
		runner := &cluster.Runner{Registry: reg}
		opts := cluster.RunOptions{
			OnlyID:          clusterPreflightCheck,
			Fix:             clusterPreflightFix,
			PerCheckTimeout: time.Duration(clusterPreflightTimeout) * time.Second,
		}
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()
		allOK, _ := runner.Run(ctx, env, opts, os.Stdout)
		if !allOK {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	clusterPreflightCmd.Flags().StringVar(&clusterPreflightCheck, "check", "", "Run only the named check (see --help for catalog)")
	clusterPreflightCmd.Flags().BoolVar(&clusterPreflightFix, "fix", false, "Attempt idempotent remediation for failing checks that support it")
	clusterPreflightCmd.Flags().StringVar(&clusterPreflightConfig, "config", "", "Path to a specific config TOML (defaults to config.$ENV_MODE.toml in cwd)")
	clusterPreflightCmd.Flags().IntVar(&clusterPreflightTimeout, "check-timeout", 10, "Per-check timeout in seconds")

	clusterCmd.AddCommand(clusterPreflightCmd)
	rootCmd.AddCommand(clusterCmd)
}
