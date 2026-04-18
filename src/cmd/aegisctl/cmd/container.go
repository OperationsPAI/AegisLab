package cmd

import (
	"fmt"
	"os"
	"strings"

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
			rows = append(rows, []string{v.Name, v.ImageRef, v.ImageRef, fmt.Sprintf("%d", v.Usage), v.UpdatedAt})
		}
		// The new IMAGE column mirrors the server-composed image_ref
		// (registry/namespace/repository:tag). Image is kept for backward
		// compatibility with existing agent scripts.
		output.PrintTable([]string{"Version", "Image", "IMAGE", "Usage", "Updated"}, rows)
		return nil
	},
}

// --- container version (subcommands) ---

var containerVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage container versions",
}

// set-image

var (
	setImageID     int
	setImageRef    string
	setImageDryRun bool
)

type setImageRequest struct {
	Registry   string `json:"registry"`
	Namespace  string `json:"namespace"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

type setImageResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Registry   string `json:"registry"`
	Namespace  string `json:"namespace"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
	ImageRef   string `json:"image_ref"`
}

var containerVersionSetImageCmd = &cobra.Command{
	Use:   "set-image",
	Short: "Rewrite the image reference of a container version in the database",
	Long: `Parse <ref> into (registry, namespace, repository, tag) and atomically
update the matching columns on the container_versions row identified by --id.

Tag is required. A reference without ":<tag>" is rejected. Digest refs
(@sha256:...) are rejected. References without a registry default to docker.io.
Nested namespaces are preserved (e.g. docker.io/foo/bar/baz:tag -> namespace
"foo/bar", repository "baz").`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if setImageID <= 0 {
			return fmt.Errorf("--id is required and must be positive")
		}
		if strings.TrimSpace(setImageRef) == "" {
			return fmt.Errorf("--ref is required")
		}
		parsed, err := parseImageRef(setImageRef)
		if err != nil {
			return err
		}

		c := newClient()

		// Fetch current row by scanning the container (versions list has id/image_ref).
		current, err := fetchContainerVersionByID(c, setImageID)
		if err != nil {
			return fmt.Errorf("failed to fetch current container version: %w", err)
		}

		if setImageDryRun {
			printSetImageDiff(current, parsed)
			if output.OutputFormat(flagOutput) == output.FormatJSON {
				output.PrintJSON(map[string]any{
					"dry_run":  true,
					"id":       setImageID,
					"current":  current,
					"proposed": parsed,
				})
			}
			return nil
		}

		req := setImageRequest{
			Registry:   parsed.Registry,
			Namespace:  parsed.Namespace,
			Repository: parsed.Repository,
			Tag:        parsed.Tag,
		}
		var resp client.APIResponse[setImageResponse]
		if err := c.Patch(fmt.Sprintf("/api/v2/container-versions/%d/image", setImageID), req, &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}
		output.PrintInfo(fmt.Sprintf("Container version %d image updated to %s", setImageID, parsed.String()))
		return nil
	},
}

// list-versions

var containerVersionListVersionsName string

var containerVersionListVersionsCmd = &cobra.Command{
	Use:   "list-versions <container-name>",
	Short: "List versions for a container (with IMAGE column)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := containerVersionListVersionsName
		if len(args) == 1 {
			name = args[0]
		}
		if name == "" {
			return fmt.Errorf("container name is required (positional arg or --name)")
		}
		return runListVersions(name)
	},
}

// fetchContainerVersionByID searches known containers for the given version id.
// Returns the version item and its parent container name.
func fetchContainerVersionByID(c *client.Client, versionID int) (*containerVersionItem, error) {
	var list client.APIResponse[client.PaginatedData[containerListItem]]
	if err := c.Get("/api/v2/containers?page=1&size=1000", &list); err != nil {
		return nil, err
	}
	r := client.NewResolver(c)
	for _, ctr := range list.Data.Items {
		id, err := r.ContainerID(ctr.Name)
		if err != nil {
			continue
		}
		var vResp client.APIResponse[client.PaginatedData[containerVersionItem]]
		if err := c.Get(fmt.Sprintf("/api/v2/containers/%d/versions?page=1&size=1000", id), &vResp); err != nil {
			continue
		}
		for _, v := range vResp.Data.Items {
			if v.ID == versionID {
				return &v, nil
			}
		}
	}
	return nil, fmt.Errorf("container version %d not found", versionID)
}

func printSetImageDiff(current *containerVersionItem, proposed imageRefParts) {
	currentRef := "(unknown)"
	if current != nil {
		currentRef = current.ImageRef
	}
	fmt.Fprintln(os.Stdout, "Dry run — no changes written.")
	fmt.Fprintf(os.Stdout, "  id:       %d\n", setImageID)
	fmt.Fprintf(os.Stdout, "  current:  %s\n", currentRef)
	fmt.Fprintf(os.Stdout, "  proposed: %s\n", proposed.String())
	fmt.Fprintf(os.Stdout, "            registry=%s namespace=%s repository=%s tag=%s\n",
		proposed.Registry, proposed.Namespace, proposed.Repository, proposed.Tag)
}

func runListVersions(containerName string) error {
	c := newClient()
	r := client.NewResolver(c)

	id, err := r.ContainerID(containerName)
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
		rows = append(rows, []string{
			fmt.Sprintf("%d", v.ID),
			v.Name,
			v.ImageRef,
			fmt.Sprintf("%d", v.Usage),
			v.UpdatedAt,
		})
	}
	// IMAGE column (registry/namespace/repository:tag) is sourced from the
	// server's image_ref field, which AfterFind composes from the four columns.
	output.PrintTable([]string{"ID", "Version", "IMAGE", "Usage", "Updated"}, rows)
	return nil
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

	// `container version` subcommands
	containerVersionSetImageCmd.Flags().IntVar(&setImageID, "id", 0, "Container version ID to update (required)")
	containerVersionSetImageCmd.Flags().StringVar(&setImageRef, "ref", "", "Full image reference <registry>/<namespace>/<repository>:<tag> (required)")
	containerVersionSetImageCmd.Flags().BoolVar(&setImageDryRun, "dry-run", false, "Print the current vs proposed diff without writing")

	containerVersionListVersionsCmd.Flags().StringVar(&containerVersionListVersionsName, "name", "", "Container name (alternatively pass as positional arg)")

	containerVersionCmd.AddCommand(containerVersionSetImageCmd)
	containerVersionCmd.AddCommand(containerVersionListVersionsCmd)
	containerCmd.AddCommand(containerVersionCmd)
}
