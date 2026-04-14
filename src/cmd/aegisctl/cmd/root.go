package cmd

import (
	"fmt"
	"os"
	"strconv"

	"aegis/cmd/aegisctl/config"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

var (
	// Global flag values.
	flagServer         string
	flagToken          string
	flagProject        string
	flagOutput         string
	flagRequestTimeout int
	flagQuiet          bool
	flagDryRun         bool

	// Resolved at PersistentPreRun time.
	cfg *config.Config
)

// rootCmd is the top-level aegisctl command.
var rootCmd = &cobra.Command{
	Use:   "aegisctl",
	Short: "CLI client for the AegisLab platform",
	Long: `aegisctl is a command-line interface for managing the AegisLab (RCABench)
fault-injection and root-cause-analysis benchmarking platform.

QUICK START:
  # 1. Login (saves token to ~/.aegisctl/config.yaml)
  aegisctl auth login --server http://HOST:8082 --username admin --password admin123

  # 2. Set default project so you don't need --project every time
  aegisctl context set --name default --default-project pair_diagnosis

  # 3. Browse available resources
  aegisctl project list
  aegisctl container list
  aegisctl container list --type algorithm
  aegisctl dataset list

  # 4. Submit a fault injection (requires a YAML spec file)
  aegisctl inject submit --spec injection.yaml --project pair_diagnosis

  # 5. Monitor progress
  aegisctl trace list --project pair_diagnosis
  aegisctl trace watch <trace-id>
  aegisctl task list --state Running
  aegisctl task logs <task-id> --follow

  # 6. Wait for completion (blocks until terminal state, exit code 0=ok, 2=fail, 3=timeout)
  aegisctl wait <trace-id> --timeout 600

  # 7. View results
  aegisctl inject list --project pair_diagnosis
  aegisctl inject get <injection-name>
  aegisctl execute list --project pair_diagnosis

OUTPUT:
  All commands support --output json (or -o json) for machine-parseable output.
  Table output goes to stdout; informational messages go to stderr.
  Use --quiet (-q) to suppress informational messages.

ENVIRONMENT VARIABLES:
  AEGIS_SERVER   - Server URL (overridden by --server flag)
  AEGIS_TOKEN    - Auth token (overridden by --token flag)
  AEGIS_PROJECT  - Default project name (overridden by --project flag)
  AEGIS_OUTPUT   - Output format: table|json (overridden by --output flag)
  AEGIS_TIMEOUT  - Request timeout in seconds (overridden by --request-timeout flag)

NAMING CONVENTION:
  Most commands accept human-readable names instead of numeric IDs.
  For example: "aegisctl container get detector" resolves "detector" to its ID.
  The --project flag also accepts project names (e.g. "pair_diagnosis").`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load configuration file.
		var err error
		cfg, err = config.LoadConfig()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		// Resolve --server: flag > env > config context.
		if flagServer == "" {
			flagServer = os.Getenv("AEGIS_SERVER")
		}
		if flagServer == "" {
			if ctx, _, err := config.GetCurrentContext(cfg); err == nil {
				flagServer = ctx.Server
			}
		}

		// Resolve --token: flag > env > config context.
		if flagToken == "" {
			flagToken = os.Getenv("AEGIS_TOKEN")
		}
		if flagToken == "" {
			if ctx, _, err := config.GetCurrentContext(cfg); err == nil {
				flagToken = ctx.Token
			}
		}

		// Resolve --project: flag > env > config context.
		if flagProject == "" {
			flagProject = os.Getenv("AEGIS_PROJECT")
		}
		if flagProject == "" {
			if ctx, _, err := config.GetCurrentContext(cfg); err == nil {
				flagProject = ctx.DefaultProject
			}
		}

		// Resolve --output: flag > env > config preferences.
		if flagOutput == "" {
			flagOutput = os.Getenv("AEGIS_OUTPUT")
		}
		if flagOutput == "" && cfg.Preferences.Output != "" {
			flagOutput = cfg.Preferences.Output
		}
		if flagOutput == "" {
			flagOutput = "table"
		}

		// Resolve --request-timeout: flag > env > config preferences.
		if flagRequestTimeout == 0 {
			if v := os.Getenv("AEGIS_TIMEOUT"); v != "" {
				if n, err := strconv.Atoi(v); err == nil {
					flagRequestTimeout = n
				}
			}
		}
		if flagRequestTimeout == 0 && cfg.Preferences.RequestTimeout > 0 {
			flagRequestTimeout = cfg.Preferences.RequestTimeout
		}
		if flagRequestTimeout == 0 {
			flagRequestTimeout = 30
		}

		// Wire quiet flag into output package.
		output.Quiet = flagQuiet

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagServer, "server", "", "AegisLab server URL (env: AEGIS_SERVER)")
	rootCmd.PersistentFlags().StringVar(&flagToken, "token", "", "Authentication token (env: AEGIS_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&flagProject, "project", "", "Default project ID (env: AEGIS_PROJECT)")
	rootCmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "Output format: table|json (env: AEGIS_OUTPUT)")
	rootCmd.PersistentFlags().IntVar(&flagRequestTimeout, "request-timeout", 0, "Request timeout in seconds (env: AEGIS_TIMEOUT)")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress informational output")
	rootCmd.PersistentFlags().BoolVar(&flagDryRun, "dry-run", false, "Show what would be done without executing")

	// Register subcommands.
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(contextCmd)

	// Stubs for subcommands implemented by other agents:
	rootCmd.AddCommand(projectCmd)
	rootCmd.AddCommand(containerCmd)
	rootCmd.AddCommand(injectCmd)
	rootCmd.AddCommand(executeCmd)
	rootCmd.AddCommand(taskCmd)
	rootCmd.AddCommand(traceCmd)
	rootCmd.AddCommand(datasetCmd)
	rootCmd.AddCommand(evalCmd)
	rootCmd.AddCommand(waitCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(completionCmd)
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
