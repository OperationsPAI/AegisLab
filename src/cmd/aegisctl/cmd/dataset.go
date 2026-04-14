package cmd

import (
	"fmt"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

// Local structs for dataset API responses.

type datasetDetail struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type datasetListItem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

type datasetVersionItem struct {
	ID        int    `json:"id"`
	Version   string `json:"version"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

var datasetCmd = &cobra.Command{
	Use:     "dataset",
	Aliases: []string{"ds"},
	Short:   "Manage datasets",
}

// --- dataset list ---

var datasetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List datasets",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		var resp client.APIResponse[client.PaginatedData[datasetListItem]]
		if err := c.Get("/api/v2/datasets?page=1&size=100", &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		rows := make([][]string, 0, len(resp.Data.Items))
		for _, d := range resp.Data.Items {
			rows = append(rows, []string{d.Name, d.Description, d.Status, d.CreatedAt})
		}
		output.PrintTable([]string{"Name", "Description", "Status", "Created"}, rows)
		return nil
	},
}

// --- dataset get ---

var datasetGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get dataset details by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		r := client.NewResolver(c)

		id, err := r.DatasetID(args[0])
		if err != nil {
			return err
		}

		var resp client.APIResponse[datasetDetail]
		if err := c.Get(fmt.Sprintf("/api/v2/datasets/%d", id), &resp); err != nil {
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

// --- dataset versions ---

var datasetVersionsCmd = &cobra.Command{
	Use:   "versions <name>",
	Short: "List versions for a dataset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		r := client.NewResolver(c)

		id, err := r.DatasetID(args[0])
		if err != nil {
			return err
		}

		var resp client.APIResponse[client.PaginatedData[datasetVersionItem]]
		if err := c.Get(fmt.Sprintf("/api/v2/datasets/%d/versions", id), &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		rows := make([][]string, 0, len(resp.Data.Items))
		for _, v := range resp.Data.Items {
			rows = append(rows, []string{v.Version, v.Status, v.CreatedAt})
		}
		output.PrintTable([]string{"Version", "Status", "Created"}, rows)
		return nil
	},
}

func init() {
	datasetCmd.AddCommand(datasetListCmd)
	datasetCmd.AddCommand(datasetGetCmd)
	datasetCmd.AddCommand(datasetVersionsCmd)
}
