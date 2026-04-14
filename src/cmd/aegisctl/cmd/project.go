package cmd

import (
	"fmt"
	"time"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

// Local structs for project API responses.

type projectDetail struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type projectListItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

func newClient() *client.Client {
	return client.NewClient(flagServer, flagToken, time.Duration(flagRequestTimeout)*time.Second)
}

func newResolver() *client.Resolver {
	return client.NewResolver(newClient())
}

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"proj"},
	Short:   "Manage projects",
}

// --- project list ---

var projectListPage int
var projectListSize int

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		path := fmt.Sprintf("/api/v2/projects?page=%d&size=%d", projectListPage, projectListSize)

		var resp client.APIResponse[client.PaginatedData[projectListItem]]
		if err := c.Get(path, &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		rows := make([][]string, 0, len(resp.Data.Items))
		for _, p := range resp.Data.Items {
			rows = append(rows, []string{p.Name, p.Description, p.Status, p.CreatedAt})
		}
		output.PrintTable([]string{"Name", "Description", "Status", "Created"}, rows)
		return nil
	},
}

// --- project get ---

var projectGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get project details by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		r := client.NewResolver(c)

		id, err := r.ProjectID(args[0])
		if err != nil {
			return err
		}

		var resp client.APIResponse[projectDetail]
		if err := c.Get(fmt.Sprintf("/api/v2/projects/%d", id), &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		fmt.Printf("Name:        %s\n", resp.Data.Name)
		fmt.Printf("ID:          %d\n", resp.Data.ID)
		fmt.Printf("Description: %s\n", resp.Data.Description)
		fmt.Printf("Status:      %s\n", resp.Data.Status)
		fmt.Printf("Created:     %s\n", resp.Data.CreatedAt)
		fmt.Printf("Updated:     %s\n", resp.Data.UpdatedAt)
		return nil
	},
}

// --- project create ---

var projectCreateName string
var projectCreateDesc string

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if projectCreateName == "" {
			return fmt.Errorf("--name is required")
		}

		c := newClient()

		body := map[string]string{
			"name":        projectCreateName,
			"description": projectCreateDesc,
		}

		var resp client.APIResponse[projectDetail]
		if err := c.Post("/api/v2/projects", body, &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		output.PrintInfo(fmt.Sprintf("Project %q created (id: %d)", resp.Data.Name, resp.Data.ID))
		return nil
	},
}

func init() {
	projectListCmd.Flags().IntVar(&projectListPage, "page", 1, "Page number")
	projectListCmd.Flags().IntVar(&projectListSize, "size", 20, "Page size")

	projectCreateCmd.Flags().StringVar(&projectCreateName, "name", "", "Project name (required)")
	projectCreateCmd.Flags().StringVar(&projectCreateDesc, "description", "", "Project description")

	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectGetCmd)
	projectCmd.AddCommand(projectCreateCmd)
}
