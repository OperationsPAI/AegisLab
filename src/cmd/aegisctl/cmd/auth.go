package cmd

import (
	"fmt"
	"time"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/config"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

// --- auth login ---

var authLoginServer string
var authLoginUsername string
var authLoginPassword string
var authLoginContext string

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with an AegisLab server",
	RunE: func(cmd *cobra.Command, args []string) error {
		server := authLoginServer
		if server == "" {
			server = flagServer
		}
		if server == "" {
			return fmt.Errorf("--server is required for login")
		}

		if authLoginUsername == "" {
			return fmt.Errorf("--username is required")
		}
		if authLoginPassword == "" {
			return fmt.Errorf("--password is required")
		}

		output.PrintInfo(fmt.Sprintf("Logging in to %s as %s...", server, authLoginUsername))

		result, err := client.Login(server, authLoginUsername, authLoginPassword)
		if err != nil {
			return err
		}

		// Determine context name.
		ctxName := authLoginContext
		if ctxName == "" {
			ctxName = "default"
		}

		// Save to config.
		cfg.Contexts[ctxName] = config.Context{
			Server:      server,
			Token:       result.Token,
			TokenExpiry: result.ExpiresAt,
		}
		cfg.CurrentContext = ctxName

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(map[string]any{
				"context":    ctxName,
				"server":     server,
				"username":   result.Username,
				"expires_at": result.ExpiresAt.Format(time.RFC3339),
			})
		} else {
			output.PrintInfo(fmt.Sprintf("Logged in as %s (context: %s)", result.Username, ctxName))
			output.PrintInfo(fmt.Sprintf("Token expires at %s", result.ExpiresAt.Format(time.RFC3339)))
		}
		return nil
	},
}

// --- auth status ---

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, ctxName, err := config.GetCurrentContext(cfg)
		if err != nil {
			return err
		}

		if ctx.Token == "" {
			return fmt.Errorf("no token set in context %q; run 'aegisctl auth login'", ctxName)
		}

		expired := client.IsTokenExpired(ctx.TokenExpiry)

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			status := "valid"
			if expired {
				status = "expired"
			}
			output.PrintJSON(map[string]any{
				"context":    ctxName,
				"server":     ctx.Server,
				"status":     status,
				"expires_at": ctx.TokenExpiry.Format(time.RFC3339),
			})
			return nil
		}

		output.PrintTable(
			[]string{"Context", "Server", "Status", "Expires"},
			[][]string{{
				ctxName,
				ctx.Server,
				func() string {
					if expired {
						return "expired"
					}
					return "valid"
				}(),
				ctx.TokenExpiry.Format(time.RFC3339),
			}},
		)

		// Also try to fetch profile to verify token is actually valid.
		profile, err := client.GetProfile(ctx.Server, ctx.Token)
		if err != nil {
			output.PrintInfo(fmt.Sprintf("Warning: could not verify token with server: %v", err))
		} else {
			output.PrintInfo(fmt.Sprintf("Authenticated as: %s (id: %d)", profile.Username, profile.ID))
		}

		return nil
	},
}

// --- auth token ---

var authTokenSet string

var authTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage authentication token",
	RunE: func(cmd *cobra.Command, args []string) error {
		if authTokenSet == "" {
			// Display current token info.
			ctx, ctxName, err := config.GetCurrentContext(cfg)
			if err != nil {
				return err
			}
			if ctx.Token == "" {
				return fmt.Errorf("no token set in context %q", ctxName)
			}
			// Show truncated token.
			token := ctx.Token
			display := token
			if len(token) > 20 {
				display = token[:10] + "..." + token[len(token)-10:]
			}
			fmt.Println(display)
			return nil
		}

		// Set token directly.
		ctxName := cfg.CurrentContext
		if ctxName == "" {
			ctxName = "default"
		}

		ctx := cfg.Contexts[ctxName]
		ctx.Token = authTokenSet
		cfg.Contexts[ctxName] = ctx
		cfg.CurrentContext = ctxName

		if err := config.SaveConfig(cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		output.PrintInfo(fmt.Sprintf("Token set for context %q", ctxName))
		return nil
	},
}

func init() {
	authLoginCmd.Flags().StringVar(&authLoginServer, "server", "", "Server URL")
	authLoginCmd.Flags().StringVar(&authLoginUsername, "username", "", "Username")
	authLoginCmd.Flags().StringVar(&authLoginPassword, "password", "", "Password")
	authLoginCmd.Flags().StringVar(&authLoginContext, "context", "", "Context name to save credentials under (default: \"default\")")

	authTokenCmd.Flags().StringVar(&authTokenSet, "set", "", "Set token directly")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authTokenCmd)
}
