package cmd

import (
	"fmt"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

// Local structs for evaluation API responses.

type evalListItem struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	ProjectID int    `json:"project_id"`
	CreatedAt string `json:"created_at"`
}

type evalDetail struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Type      string `json:"type"`
	ProjectID int    `json:"project_id"`
	Result    any    `json:"result"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

var evalCmd = &cobra.Command{
	Use:     "eval",
	Aliases: []string{"evaluation"},
	Short:   "Manage evaluations",
}

// --- eval list ---

var evalListPage int
var evalListSize int

var evalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List evaluations",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		path := fmt.Sprintf("/api/v2/evaluations?page=%d&size=%d", evalListPage, evalListSize)

		var resp client.APIResponse[client.PaginatedData[evalListItem]]
		if err := c.Get(path, &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		rows := make([][]string, 0, len(resp.Data.Items))
		for _, e := range resp.Data.Items {
			rows = append(rows, []string{
				fmt.Sprintf("%d", e.ID),
				e.Name,
				e.Type,
				e.Status,
				e.CreatedAt,
			})
		}
		output.PrintTable([]string{"ID", "Name", "Type", "Status", "Created"}, rows)
		return nil
	},
}

// --- eval get ---

var evalGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get evaluation details by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		var resp client.APIResponse[evalDetail]
		if err := c.Get(fmt.Sprintf("/api/v2/evaluations/%s", args[0]), &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		fmt.Printf("ID:        %d\n", resp.Data.ID)
		fmt.Printf("Name:      %s\n", resp.Data.Name)
		fmt.Printf("Type:      %s\n", resp.Data.Type)
		fmt.Printf("Status:    %s\n", resp.Data.Status)
		fmt.Printf("Project:   %d\n", resp.Data.ProjectID)
		fmt.Printf("Created:   %s\n", resp.Data.CreatedAt)
		fmt.Printf("Updated:   %s\n", resp.Data.UpdatedAt)
		if resp.Data.Result != nil {
			fmt.Printf("Result:\n")
			output.PrintJSON(resp.Data.Result)
		}
		return nil
	},
}

func init() {
	evalListCmd.Flags().IntVar(&evalListPage, "page", 1, "Page number")
	evalListCmd.Flags().IntVar(&evalListSize, "size", 20, "Page size")

	evalCmd.AddCommand(evalListCmd)
	evalCmd.AddCommand(evalGetCmd)
}
