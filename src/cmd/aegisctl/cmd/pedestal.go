package cmd

import (
	"fmt"
	"os"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
)

// Local types mirroring the backend response shapes in src/dto/container.go.
// We duplicate these instead of importing aegis/dto to keep the aegisctl
// binary free of the full server's gorm+entity graph.

type pedestalHelmConfig struct {
	ID                 int    `json:"id"`
	ContainerVersionID int    `json:"container_version_id"`
	ChartName          string `json:"chart_name"`
	Version            string `json:"version"`
	RepoURL            string `json:"repo_url"`
	RepoName           string `json:"repo_name"`
	ValueFile          string `json:"value_file"`
	LocalPath          string `json:"local_path"`
	Checksum           string `json:"checksum,omitempty"`
}

type pedestalHelmVerifyCheck struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

type pedestalHelmVerifyResp struct {
	OK     bool                      `json:"ok"`
	Checks []pedestalHelmVerifyCheck `json:"checks"`
}

type pedestalHelmSetReq struct {
	ChartName string `json:"chart_name"`
	Version   string `json:"version"`
	RepoURL   string `json:"repo_url"`
	RepoName  string `json:"repo_name"`
	ValueFile string `json:"value_file,omitempty"`
	LocalPath string `json:"local_path,omitempty"`
}

var pedestalCmd = &cobra.Command{
	Use:   "pedestal",
	Short: "Manage pedestal (SUT) container infrastructure",
}

var pedestalHelmCmd = &cobra.Command{
	Use:   "helm",
	Short: "Inspect and edit helm_configs rows bound to pedestal container versions",
	Long: `Manage the helm chart configuration bound to a pedestal container version.

Typical workflow for fixing a bad repo URL without touching MySQL directly:

  aegisctl pedestal helm get    --container-version-id 42
  aegisctl pedestal helm set    --container-version-id 42 --repo-url https://...
  aegisctl pedestal helm verify --container-version-id 42`,
}

// --- shared flags ---
var pedestalHelmVersionID int

// --- get ---
var pedestalHelmGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the helm_configs row for a container version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if pedestalHelmVersionID <= 0 {
			return fmt.Errorf("--container-version-id is required and must be > 0")
		}
		c := newClient()
		var resp client.APIResponse[pedestalHelmConfig]
		if err := c.Get(fmt.Sprintf("/api/v2/pedestal/helm/%d", pedestalHelmVersionID), &resp); err != nil {
			return err
		}
		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}
		d := resp.Data
		fmt.Printf("ID:                   %d\n", d.ID)
		fmt.Printf("ContainerVersionID:   %d\n", d.ContainerVersionID)
		fmt.Printf("ChartName:            %s\n", d.ChartName)
		fmt.Printf("Version:              %s\n", d.Version)
		fmt.Printf("RepoURL:              %s\n", d.RepoURL)
		fmt.Printf("RepoName:             %s\n", d.RepoName)
		fmt.Printf("ValueFile:            %s\n", d.ValueFile)
		fmt.Printf("LocalPath:            %s\n", d.LocalPath)
		if d.Checksum != "" {
			fmt.Printf("Checksum:             %s\n", d.Checksum)
		}
		return nil
	},
}

// --- set ---
var (
	pedestalHelmSetChart    string
	pedestalHelmSetVersion  string
	pedestalHelmSetRepoURL  string
	pedestalHelmSetRepoName string
	pedestalHelmSetValues   string
	pedestalHelmSetLocal    string
)

var pedestalHelmSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Upsert the helm_configs row for a container version (admin only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if pedestalHelmVersionID <= 0 {
			return fmt.Errorf("--container-version-id is required and must be > 0")
		}
		if pedestalHelmSetChart == "" || pedestalHelmSetVersion == "" ||
			pedestalHelmSetRepoURL == "" || pedestalHelmSetRepoName == "" {
			return fmt.Errorf("--chart-name, --version, --repo-url, --repo-name are all required")
		}
		body := pedestalHelmSetReq{
			ChartName: pedestalHelmSetChart,
			Version:   pedestalHelmSetVersion,
			RepoURL:   pedestalHelmSetRepoURL,
			RepoName:  pedestalHelmSetRepoName,
			ValueFile: pedestalHelmSetValues,
			LocalPath: pedestalHelmSetLocal,
		}
		c := newClient()
		var resp client.APIResponse[pedestalHelmConfig]
		if err := c.Put(fmt.Sprintf("/api/v2/pedestal/helm/%d", pedestalHelmVersionID), body, &resp); err != nil {
			return err
		}
		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}
		output.PrintInfo(fmt.Sprintf("helm_configs row for container_version_id=%d upserted (id=%d)",
			resp.Data.ContainerVersionID, resp.Data.ID))
		return nil
	},
}

// --- verify ---
var pedestalHelmVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Dry-run helm repo add + pull + value-file parse without starting a restart task",
	RunE: func(cmd *cobra.Command, args []string) error {
		if pedestalHelmVersionID <= 0 {
			return fmt.Errorf("--container-version-id is required and must be > 0")
		}
		c := newClient()
		var resp client.APIResponse[pedestalHelmVerifyResp]
		if err := c.Post(fmt.Sprintf("/api/v2/pedestal/helm/%d/verify", pedestalHelmVersionID), nil, &resp); err != nil {
			return err
		}
		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			if !resp.Data.OK {
				os.Exit(1)
			}
			return nil
		}
		for _, chk := range resp.Data.Checks {
			mark := "OK"
			if !chk.OK {
				mark = "FAIL"
			}
			fmt.Printf("[%s] %s\n", mark, chk.Name)
			if chk.Detail != "" {
				fmt.Printf("       %s\n", chk.Detail)
			}
		}
		if !resp.Data.OK {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	// Shared flag across all three subcommands.
	for _, c := range []*cobra.Command{pedestalHelmGetCmd, pedestalHelmSetCmd, pedestalHelmVerifyCmd} {
		c.Flags().IntVar(&pedestalHelmVersionID, "container-version-id", 0, "Container version ID (required)")
		_ = c.MarkFlagRequired("container-version-id")
	}

	pedestalHelmSetCmd.Flags().StringVar(&pedestalHelmSetChart, "chart-name", "", "Helm chart name (required)")
	pedestalHelmSetCmd.Flags().StringVar(&pedestalHelmSetVersion, "version", "", "Helm chart version, semver (required)")
	pedestalHelmSetCmd.Flags().StringVar(&pedestalHelmSetRepoURL, "repo-url", "", "Helm repository URL (required)")
	pedestalHelmSetCmd.Flags().StringVar(&pedestalHelmSetRepoName, "repo-name", "", "Helm repository name / alias (required)")
	pedestalHelmSetCmd.Flags().StringVar(&pedestalHelmSetValues, "values-file", "", "Path to the values YAML file (optional)")
	pedestalHelmSetCmd.Flags().StringVar(&pedestalHelmSetLocal, "local-path", "", "Local chart fallback path (optional)")

	pedestalHelmCmd.AddCommand(pedestalHelmGetCmd)
	pedestalHelmCmd.AddCommand(pedestalHelmSetCmd)
	pedestalHelmCmd.AddCommand(pedestalHelmVerifyCmd)

	pedestalCmd.AddCommand(pedestalHelmCmd)
}
