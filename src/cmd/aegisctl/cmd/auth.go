package cmd

import (
	"fmt"
	"os"
	"strings"
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
var authLoginKeyID string
var authLoginKeySecret string
var authLoginContext string

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Exchange Key ID / Key Secret for a bearer token",
	RunE: func(cmd *cobra.Command, args []string) error {
		server := authLoginServer
		if server == "" {
			server = flagServer
		}
		if server == "" {
			return fmt.Errorf("--server is required for login")
		}

		keyID := authLoginKeyID
		if keyID == "" {
			keyID = os.Getenv("AEGIS_KEY_ID")
		}
		if keyID == "" {
			return fmt.Errorf("--key-id is required")
		}

		keySecret := authLoginKeySecret
		if keySecret == "" {
			keySecret = os.Getenv("AEGIS_KEY_SECRET")
		}
		if keySecret == "" {
			return fmt.Errorf("--key-secret is required")
		}

		output.PrintInfo(fmt.Sprintf("Exchanging API key token with %s using %s...", server, keyID))

		result, err := client.LoginWithAPIKey(server, keyID, keySecret)
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
			AuthType:    result.AuthType,
			KeyID:       result.KeyID,
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
				"auth_type":  result.AuthType,
				"key_id":     result.KeyID,
				"expires_at": result.ExpiresAt.Format(time.RFC3339),
			})
		} else {
			output.PrintInfo(fmt.Sprintf("Token issued for key id %s (context: %s)", result.KeyID, ctxName))
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
				"auth_type":  ctx.AuthType,
				"key_id":     ctx.KeyID,
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
		if ctx.KeyID != "" {
			output.PrintInfo(fmt.Sprintf("Issued via key id: %s", ctx.KeyID))
		}

		return nil
	},
}

// --- auth inspect ---

var authInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect locally stored authentication context",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, ctxName, err := config.GetCurrentContext(cfg)
		if err != nil {
			return err
		}

		tokenPreview := ""
		if ctx.Token != "" {
			tokenPreview = ctx.Token
			if len(tokenPreview) > 20 {
				tokenPreview = tokenPreview[:10] + "..." + tokenPreview[len(tokenPreview)-10:]
			}
		}
		expiresAt := ""
		if !ctx.TokenExpiry.IsZero() {
			expiresAt = ctx.TokenExpiry.Format(time.RFC3339)
		}
		expired := false
		if !ctx.TokenExpiry.IsZero() {
			expired = client.IsTokenExpired(ctx.TokenExpiry)
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(map[string]any{
				"context":       ctxName,
				"server":        ctx.Server,
				"auth_type":     ctx.AuthType,
				"key_id":        ctx.KeyID,
				"token_present": ctx.Token != "",
				"token_preview": tokenPreview,
				"token_expired": expired,
				"expires_at":    expiresAt,
			})
			return nil
		}

		output.PrintTable(
			[]string{"Context", "Server", "AuthType", "KeyID", "Token", "Expired", "Expires"},
			[][]string{{
				ctxName,
				ctx.Server,
				emptyOrValue(ctx.AuthType, "-"),
				emptyOrValue(ctx.KeyID, "-"),
				emptyOrValue(tokenPreview, "-"),
				fmt.Sprintf("%t", expired),
				emptyOrValue(expiresAt, "-"),
			}},
		)
		return nil
	},
}

// --- auth sign-debug ---

var authSignDebugKeyID string
var authSignDebugKeySecret string
var authSignDebugTimestamp int64
var authSignDebugNonce string
var authSignDebugExecute bool
var authSignDebugSaveContext bool

var authSignDebugCmd = &cobra.Command{
	Use:   "sign-debug",
	Short: "Print canonical string and signature headers for Key ID / Key Secret token exchange",
	RunE: func(cmd *cobra.Command, args []string) error {
		keyID := authSignDebugKeyID
		if keyID == "" {
			keyID = os.Getenv("AEGIS_KEY_ID")
		}
		if keyID == "" {
			return fmt.Errorf("--key-id is required")
		}

		keySecret := authSignDebugKeySecret
		if keySecret == "" {
			keySecret = os.Getenv("AEGIS_KEY_SECRET")
		}
		if keySecret == "" {
			return fmt.Errorf("--key-secret is required")
		}

		signTime := time.Now().UTC()
		if authSignDebugTimestamp > 0 {
			signTime = time.Unix(authSignDebugTimestamp, 0).UTC()
		}

		debugInfo, err := client.PrepareAPIKeyTokenDebug(keyID, keySecret, signTime, authSignDebugNonce)
		if err != nil {
			return err
		}

		server := strings.TrimRight(resolveServerForAuthDebug(), "/")
		if authSignDebugExecute && (server == "" || strings.Contains(server, "HOST:8082")) {
			return fmt.Errorf("--execute requires a real --server or configured AEGIS_SERVER/current context")
		}
		curlCommand := buildAPIKeyCurl(server, debugInfo)
		var executeResp map[string]any
		if authSignDebugExecute {
			executeResp, err = executeAPIKeyTokenExchange(server, debugInfo)
			if err != nil {
				return err
			}
			if authSignDebugSaveContext {
				if err := saveAPIKeyContext(server, executeResp); err != nil {
					return err
				}
			}
		} else if authSignDebugSaveContext {
			return fmt.Errorf("--save-context requires --execute")
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			result := map[string]any{
				"server":           server,
				"method":           debugInfo.Method,
				"path":             debugInfo.Path,
				"key_id":           debugInfo.KeyID,
				"timestamp":        debugInfo.Timestamp,
				"nonce":            debugInfo.Nonce,
				"body_sha256":      debugInfo.BodySHA256,
				"canonical_string": debugInfo.CanonicalString,
				"signature":        debugInfo.Signature,
				"headers":          debugInfo.Headers(),
				"curl":             curlCommand,
				"executed":         authSignDebugExecute,
				"saved_context":    authSignDebugSaveContext,
			}
			if authSignDebugExecute {
				result["response"] = executeResp
			}
			output.PrintJSON(result)
			return nil
		}

		fmt.Printf("Server: %s\n", server)
		fmt.Printf("Method: %s\n", debugInfo.Method)
		fmt.Printf("Path: %s\n", debugInfo.Path)
		fmt.Printf("Key-Id: %s\n", debugInfo.KeyID)
		fmt.Printf("Timestamp: %s\n", debugInfo.Timestamp)
		fmt.Printf("Nonce: %s\n", debugInfo.Nonce)
		fmt.Printf("Body-SHA256: %s\n", debugInfo.BodySHA256)
		fmt.Printf("Signature: %s\n\n", debugInfo.Signature)
		fmt.Println("Canonical String:")
		fmt.Println(debugInfo.CanonicalString)
		fmt.Println()
		fmt.Println("curl:")
		fmt.Println(curlCommand)
		if authSignDebugExecute {
			fmt.Println()
			fmt.Println("response:")
			output.PrintJSON(executeResp)
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
		ctx.AuthType = "token"
		ctx.KeyID = ""
		ctx.TokenExpiry = time.Time{}
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
	authLoginCmd.Flags().StringVar(&authLoginKeyID, "key-id", "", "Key ID (env: AEGIS_KEY_ID)")
	authLoginCmd.Flags().StringVar(&authLoginKeySecret, "key-secret", "", "Key secret (env: AEGIS_KEY_SECRET)")
	authLoginCmd.Flags().StringVar(&authLoginContext, "context", "", "Context name to save credentials under (default: \"default\")")
	authSignDebugCmd.Flags().StringVar(&authSignDebugKeyID, "key-id", "", "Key ID (env: AEGIS_KEY_ID)")
	authSignDebugCmd.Flags().StringVar(&authSignDebugKeySecret, "key-secret", "", "Key secret (env: AEGIS_KEY_SECRET)")
	authSignDebugCmd.Flags().Int64Var(&authSignDebugTimestamp, "timestamp", 0, "Override unix timestamp in seconds")
	authSignDebugCmd.Flags().StringVar(&authSignDebugNonce, "nonce", "", "Override nonce for reproducible signature output")
	authSignDebugCmd.Flags().BoolVar(&authSignDebugExecute, "execute", false, "Execute the signed token exchange request and print the response")
	authSignDebugCmd.Flags().BoolVar(&authSignDebugSaveContext, "save-context", false, "Save the exchanged bearer token into the current context after --execute succeeds")

	authTokenCmd.Flags().StringVar(&authTokenSet, "set", "", "Set token directly")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authInspectCmd)
	authCmd.AddCommand(authSignDebugCmd)
	authCmd.AddCommand(authTokenCmd)
}

func resolveServerForAuthDebug() string {
	if flagServer != "" {
		return flagServer
	}
	if value := os.Getenv("AEGIS_SERVER"); value != "" {
		return value
	}
	if cfg != nil {
		if ctx, _, err := config.GetCurrentContext(cfg); err == nil && ctx.Server != "" {
			return ctx.Server
		}
	}
	return "http://HOST:8082"
}

func buildAPIKeyCurl(server string, debugInfo *client.APIKeyTokenDebug) string {
	return fmt.Sprintf(
		"curl -X POST %s%s -H 'Accept: application/json' -H 'X-Key-Id: %s' -H 'X-Timestamp: %s' -H 'X-Nonce: %s' -H 'X-Signature: %s'",
		server,
		debugInfo.Path,
		debugInfo.KeyID,
		debugInfo.Timestamp,
		debugInfo.Nonce,
		debugInfo.Signature,
	)
}

func executeAPIKeyTokenExchange(server string, debugInfo *client.APIKeyTokenDebug) (map[string]any, error) {
	httpClient := client.NewClient(server, "", 30*time.Second)
	var response map[string]any
	if err := httpClient.PostWithHeaders(debugInfo.Path, debugInfo.Headers(), &response); err != nil {
		return nil, fmt.Errorf("execute token exchange: %w", err)
	}
	return response, nil
}

func saveAPIKeyContext(server string, executeResp map[string]any) error {
	ctxName := resolveContextNameForSave()
	ctx := cfg.Contexts[ctxName]
	ctx.Server = server

	data, ok := executeResp["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("execute response does not contain a valid data payload")
	}

	token, _ := data["token"].(string)
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("execute response does not contain a token")
	}
	ctx.Token = token

	if authType, _ := data["auth_type"].(string); strings.TrimSpace(authType) != "" {
		ctx.AuthType = authType
	}
	if keyID, _ := data["key_id"].(string); strings.TrimSpace(keyID) != "" {
		ctx.KeyID = keyID
	}
	if expiresAt, _ := data["expires_at"].(string); strings.TrimSpace(expiresAt) != "" {
		parsed, err := time.Parse(time.RFC3339, expiresAt)
		if err != nil {
			return fmt.Errorf("parse expires_at: %w", err)
		}
		ctx.TokenExpiry = parsed
	}

	cfg.Contexts[ctxName] = ctx
	cfg.CurrentContext = ctxName
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	output.PrintInfo(fmt.Sprintf("Saved token to context %q", ctxName))
	return nil
}

func resolveContextNameForSave() string {
	if cfg != nil && strings.TrimSpace(cfg.CurrentContext) != "" {
		return cfg.CurrentContext
	}
	return "default"
}

func emptyOrValue(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
