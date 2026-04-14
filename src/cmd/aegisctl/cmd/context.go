package cmd

import (
	"fmt"

	"aegis/cmd/aegisctl/config"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage connection contexts",
}

// --- context set ---

var contextSetName string
var contextSetServer string
var contextSetDefaultProject string

var contextSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Create or update a context",
	RunE: func(cmd *cobra.Command, args []string) error {
		if contextSetName == "" {
			return fmt.Errorf("--name is required")
		}

		ctx, exists := cfg.Contexts[contextSetName]
		if !exists {
			ctx = config.Context{}
		}

		if contextSetServer != "" {
			ctx.Server = contextSetServer
		}
		if contextSetDefaultProject != "" {
			ctx.DefaultProject = contextSetDefaultProject
		}

		cfg.Contexts[contextSetName] = ctx

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		action := "Updated"
		if !exists {
			action = "Created"
		}
		output.PrintInfo(fmt.Sprintf("%s context %q", action, contextSetName))
		return nil
	},
}

// --- context use ---

var contextUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch to a named context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if err := config.SetCurrentContext(cfg, name); err != nil {
			return err
		}
		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		output.PrintInfo(fmt.Sprintf("Switched to context %q", name))
		return nil
	},
}

// --- context list ---

var contextListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(cfg.Contexts) == 0 {
			output.PrintInfo("No contexts configured. Run 'aegisctl auth login' or 'aegisctl context set'.")
			return nil
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(cfg.Contexts)
			return nil
		}

		headers := []string{"Current", "Name", "Server", "Project"}
		var rows [][]string
		for name, ctx := range cfg.Contexts {
			current := ""
			if name == cfg.CurrentContext {
				current = "*"
			}
			rows = append(rows, []string{current, name, ctx.Server, ctx.DefaultProject})
		}

		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	contextSetCmd.Flags().StringVar(&contextSetName, "name", "", "Context name (required)")
	contextSetCmd.Flags().StringVar(&contextSetServer, "server", "", "Server URL")
	contextSetCmd.Flags().StringVar(&contextSetDefaultProject, "default-project", "", "Default project ID")

	contextCmd.AddCommand(contextSetCmd)
	contextCmd.AddCommand(contextUseCmd)
	contextCmd.AddCommand(contextListCmd)
}
