package cmd

import (
	"fmt"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

// Local structs for container API responses.

type containerDetail struct {
	ID        int                    `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Versions  []containerVersionItem `json:"versions"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type containerGetOutput struct {
	containerDetail
	DefaultVersion string `json:"default_version"`
	VersionCount   int    `json:"version_count"`
}

type containerListItem struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type containerVersionItem struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ImageRef  string `json:"image_ref"`
	Usage     int    `json:"usage"`
	UpdatedAt string `json:"updated_at"`
}

var containerCmd = &cobra.Command{
	Use:     "container",
	Aliases: []string{"ctr"},
	Short:   "Manage containers",
}

// --- container list ---

var containerListType string

// containerTypeNameToInt converts a human-readable container type to its API integer.
var containerTypeNameToInt = map[string]string{
	"algorithm": "0",
	"benchmark": "1",
	"pedestal":  "2",
}

var containerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		path := "/api/v2/containers?page=1&size=100"
		if containerListType != "" {
			typeInt, ok := containerTypeNameToInt[containerListType]
			if !ok {
				return fmt.Errorf("invalid container type %q (valid: algorithm, benchmark, pedestal)", containerListType)
			}
			path += "&type=" + typeInt
		}

		var resp client.APIResponse[client.PaginatedData[containerListItem]]
		if err := c.Get(path, &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		rows := make([][]string, 0, len(resp.Data.Items))
		for _, item := range resp.Data.Items {
			rows = append(rows, []string{item.Name, item.Type, item.Status, item.CreatedAt})
		}
		output.PrintTable([]string{"Name", "Type", "Status", "Created"}, rows)
		return nil
	},
}

// --- container get ---

var containerGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get container details by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		r := client.NewResolver(c)

		id, err := r.ContainerID(args[0])
		if err != nil {
			return err
		}

		var resp client.APIResponse[containerDetail]
		if err := c.Get(fmt.Sprintf("/api/v2/containers/%d", id), &resp); err != nil {
			return err
		}

		versionCount := len(resp.Data.Versions)
		defaultVersion := "(none)"
		if versionCount > 0 {
			defaultVersion = resp.Data.Versions[0].Name
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			out := containerGetOutput{
				containerDetail: resp.Data,
				DefaultVersion:  defaultVersion,
				VersionCount:    versionCount,
			}
			output.PrintJSON(out)
			return nil
		}

		fmt.Printf("Name:     %s\n", resp.Data.Name)
		fmt.Printf("ID:       %d\n", resp.Data.ID)
		fmt.Printf("Type:     %s\n", resp.Data.Type)
		fmt.Printf("Status:   %s\n", resp.Data.Status)
		fmt.Printf("Versions: %d\n", versionCount)
		fmt.Printf("Default:  %s\n", defaultVersion)
		fmt.Printf("Created:  %s\n", resp.Data.CreatedAt)
		fmt.Printf("Updated:  %s\n", resp.Data.UpdatedAt)
		return nil
	},
}

// --- container versions ---

var containerVersionsCmd = &cobra.Command{
	Use:   "versions <name>",
	Short: "List versions for a container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		r := client.NewResolver(c)

		id, err := r.ContainerID(args[0])
		if err != nil {
			return err
		}

		var resp client.APIResponse[client.PaginatedData[containerVersionItem]]
		if err := c.Get(fmt.Sprintf("/api/v2/containers/%d/versions", id), &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		rows := make([][]string, 0, len(resp.Data.Items))
		for _, v := range resp.Data.Items {
			rows = append(rows, []string{v.Name, v.ImageRef, fmt.Sprintf("%d", v.Usage), v.UpdatedAt})
		}
		output.PrintTable([]string{"Version", "Image", "Usage", "Updated"}, rows)
		return nil
	},
}

// --- container build ---

var containerBuildVersion string

var containerBuildCmd = &cobra.Command{
	Use:   "build <name>",
	Short: "Trigger a container build",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		body := map[string]string{
			"name": args[0],
		}
		if containerBuildVersion != "" {
			body["version"] = containerBuildVersion
		}

		var resp client.APIResponse[any]
		if err := c.Post("/api/v2/containers/build", body, &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		output.PrintInfo(fmt.Sprintf("Build triggered for container %q", args[0]))
		return nil
	},
}

func init() {
	containerListCmd.Flags().StringVar(&containerListType, "type", "", "Filter by type: algorithm|benchmark|pedestal")

	containerBuildCmd.Flags().StringVar(&containerBuildVersion, "version", "", "Version tag for the build")

	containerCmd.AddCommand(containerListCmd)
	containerCmd.AddCommand(containerGetCmd)
	containerCmd.AddCommand(containerVersionsCmd)
	containerCmd.AddCommand(containerBuildCmd)
}
