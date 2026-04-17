package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"aegis/cmd/aegisctl/client"
	"aegis/cmd/aegisctl/output"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ---------- spec types ----------

// InjectSpec is the YAML structure for injection submission.
type InjectSpec struct {
	Pedestal    ContainerRef   `yaml:"pedestal"    json:"pedestal"`
	Benchmark   ContainerRef   `yaml:"benchmark"   json:"benchmark"`
	Interval    int            `yaml:"interval"     json:"interval"`
	PreDuration int            `yaml:"pre_duration" json:"pre_duration"`
	Specs       [][]FaultSpec  `yaml:"specs"        json:"specs"`
	Algorithms  []ContainerRef `yaml:"algorithms,omitempty" json:"algorithms,omitempty"`
	Labels      []LabelItem    `yaml:"labels,omitempty"     json:"labels,omitempty"`
}

// ContainerRef references a container image with optional overrides.
type ContainerRef struct {
	Name    string          `yaml:"name"                 json:"name"`
	Version string          `yaml:"version,omitempty"    json:"version,omitempty"`
	EnvVars []ParameterSpec `yaml:"env_vars,omitempty"   json:"env_vars,omitempty"`
	Payload map[string]any  `yaml:"payload,omitempty"    json:"payload,omitempty"`
}

// FaultSpec describes a single fault to inject.
type FaultSpec struct {
	Type      string         `yaml:"type"      json:"type"`
	Namespace string         `yaml:"namespace" json:"namespace"`
	Target    string         `yaml:"target"    json:"target"`
	Duration  string         `yaml:"duration"  json:"duration"`
	Params    map[string]any `yaml:"params,omitempty" json:"params,omitempty"` // Additional spec-specific parameters (e.g., cpu_load, cpu_worker)
}

// LabelItem is a key-value label.
type LabelItem struct {
	Key   string `yaml:"key"   json:"key"`
	Value string `yaml:"value" json:"value"`
}

// ParameterSpec is a key-value parameter (env var, etc.).
type ParameterSpec struct {
	Key   string `yaml:"key"   json:"key"`
	Value string `yaml:"value" json:"value"`
}

// ---------- helpers ----------

func requireProjectName() (string, error) {
	if flagProject == "" {
		return "", fmt.Errorf("--project is required")
	}
	return flagProject, nil
}

func resolveProjectIDByName() (int, error) {
	name, err := requireProjectName()
	if err != nil {
		return 0, err
	}
	return newResolver().ProjectID(name)
}

// ---------- inject root ----------

var injectCmd = &cobra.Command{
	Use:   "inject",
	Short: "Manage fault injections",
	Long: `Manage fault injections in AegisLab projects.

WORKFLOW:
  # Submit injection from a YAML spec file
  aegisctl inject submit --spec injection.yaml --project pair_diagnosis

  # List injections in a project
  aegisctl inject list --project pair_diagnosis

  # Get details of a specific injection by name
  aegisctl inject get <injection-name>

  # Search injections by pattern
  aegisctl inject search --name-pattern "cpu*" --project pair_diagnosis

  # View injection logs and files
  aegisctl inject logs <injection-name>
  aegisctl inject files <injection-name>

  # Download injection artifacts
  aegisctl inject download <injection-name> -O ./output.tar.gz

  # View fault type metadata (available fault types and their IDs)
  aegisctl inject metadata

SPEC FILE FORMAT (injection.yaml):
  pedestal:
    name: otel-demo
    version: "1.0.0"
  benchmark:
    name: clickhouse
    version: "1.0.0"
  interval: 60
  pre_duration: 30
  algorithms:
    - name: random
      version: "1.0.0"
  specs:
    - - type: CPUStress
        namespace: exp
        target: frontend
        duration: "5m"
        params:
          cpu_load: 80
          cpu_worker: 2
  labels:
    - key: experiment
      value: cpu-stress-test

SUPPORTED FAULT TYPES:
  Run 'aegisctl inject metadata' to see all available fault types and their parameters.
  Common types: CPUStress, MemoryStress, PodKill, PodFailure, ContainerKill,
  HTTPRequestAbort, HTTPResponseDelay, NetworkDelay, NetworkLoss, DNSError, etc.

NOTE: --project is required for submit, list, and search commands.
      It accepts project names (resolved to IDs automatically).
      The 'target' field accepts container names (resolved to indices automatically).
      Duration accepts Go time strings: "60s", "5m", "1h", etc.`,
}

// ---------- inject submit ----------

var injectSubmitSpec string

var injectSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit a fault injection from a YAML spec",
	RunE: func(cmd *cobra.Command, args []string) error {
		if injectSubmitSpec == "" {
			return fmt.Errorf("--spec is required")
		}

		data, err := os.ReadFile(injectSubmitSpec)
		if err != nil {
			return fmt.Errorf("read spec file: %w", err)
		}

		var spec InjectSpec
		if err := yaml.Unmarshal(data, &spec); err != nil {
			return fmt.Errorf("parse spec YAML: %w", err)
		}

		pid, err := resolveProjectIDByName()
		if err != nil {
			return err
		}

		c := newClient()
		path := fmt.Sprintf("/api/v2/projects/%d/injections/inject", pid)

		var resp client.APIResponse[any]
		if err := c.Post(path, spec, &resp); err != nil {
			return err
		}

		output.PrintJSON(resp.Data)
		return nil
	},
}

// ---------- inject list ----------

var (
	injectListState     string
	injectListFaultType string
	injectListLabels    string
	injectListPage      int
	injectListSize      int
)

var injectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List fault injections in a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := resolveProjectIDByName()
		if err != nil {
			return err
		}

		c := newClient()
		q := fmt.Sprintf("/api/v2/projects/%d/injections?page=%d&size=%d", pid, injectListPage, injectListSize)
		if injectListState != "" {
			q += "&state=" + injectListState
		}
		if injectListFaultType != "" {
			q += "&fault_type=" + injectListFaultType
		}
		if injectListLabels != "" {
			q += "&labels=" + injectListLabels
		}

		type listItem struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			State     string `json:"state"`
			FaultType string `json:"fault_type"`
			StartTime string `json:"start_time"`
			Labels    []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			} `json:"labels"`
		}

		var resp client.APIResponse[client.PaginatedData[listItem]]
		if err := c.Get(q, &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		headers := []string{"NAME", "STATE", "FAULT-TYPE", "START-TIME", "LABELS"}
		var rows [][]string
		for _, item := range resp.Data.Items {
			var lbls []string
			for _, l := range item.Labels {
				lbls = append(lbls, l.Key+"="+l.Value)
			}
			rows = append(rows, []string{
				item.Name,
				item.State,
				item.FaultType,
				item.StartTime,
				strings.Join(lbls, ","),
			})
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

// ---------- inject get ----------

var injectGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get detailed info about an injection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r := newResolver()
		id, err := r.InjectionID(args[0])
		if err != nil {
			return err
		}

		c := newClient()
		var resp client.APIResponse[any]
		if err := c.Get(fmt.Sprintf("/api/v2/injections/%d", id), &resp); err != nil {
			return err
		}

		output.PrintJSON(resp.Data)
		return nil
	},
}

// ---------- inject search ----------

var (
	injectSearchNamePattern string
	injectSearchLabels      string
)

var injectSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search injections in a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := resolveProjectIDByName()
		if err != nil {
			return err
		}

		body := map[string]any{}
		if injectSearchNamePattern != "" {
			body["name_pattern"] = injectSearchNamePattern
		}
		if injectSearchLabels != "" {
			body["labels"] = injectSearchLabels
		}

		c := newClient()
		var resp client.APIResponse[any]
		if err := c.Post(fmt.Sprintf("/api/v2/projects/%d/injections/search", pid), body, &resp); err != nil {
			return err
		}

		output.PrintJSON(resp.Data)
		return nil
	},
}

// ---------- inject logs ----------

var injectLogsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "Show logs for an injection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r := newResolver()
		id, err := r.InjectionID(args[0])
		if err != nil {
			return err
		}

		c := newClient()
		var resp client.APIResponse[any]
		if err := c.Get(fmt.Sprintf("/api/v2/injections/%d/logs", id), &resp); err != nil {
			return err
		}

		output.PrintJSON(resp.Data)
		return nil
	},
}

// ---------- inject files ----------

var injectFilesCmd = &cobra.Command{
	Use:   "files <name>",
	Short: "List files produced by an injection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r := newResolver()
		id, err := r.InjectionID(args[0])
		if err != nil {
			return err
		}

		type fileItem struct {
			Path string `json:"path"`
			Size string `json:"size"`
			Type string `json:"type"`
		}

		c := newClient()
		var resp client.APIResponse[[]fileItem]
		if err := c.Get(fmt.Sprintf("/api/v2/injections/%d/files", id), &resp); err != nil {
			return err
		}

		if output.OutputFormat(flagOutput) == output.FormatJSON {
			output.PrintJSON(resp.Data)
			return nil
		}

		headers := []string{"PATH", "SIZE", "TYPE"}
		var rows [][]string
		for _, f := range resp.Data {
			rows = append(rows, []string{f.Path, f.Size, f.Type})
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

// ---------- inject download ----------

var injectDownloadOutput string

var injectDownloadCmd = &cobra.Command{
	Use:   "download <name>",
	Short: "Download injection artifacts to a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if injectDownloadOutput == "" {
			return fmt.Errorf("-o <path> is required")
		}

		r := newResolver()
		id, err := r.InjectionID(args[0])
		if err != nil {
			return err
		}

		// Build a raw HTTP request (binary download, not JSON).
		url := flagServer + fmt.Sprintf("/api/v2/injections/%d/download", id)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		if flagToken != "" {
			req.Header.Set("Authorization", "Bearer "+flagToken)
		}

		httpClient := &http.Client{Timeout: time.Duration(flagRequestTimeout) * time.Second}
		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("download request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("download failed (HTTP %d): %s", resp.StatusCode, string(body))
		}

		f, err := os.Create(injectDownloadOutput)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer f.Close()

		n, err := io.Copy(f, resp.Body)
		if err != nil {
			return fmt.Errorf("write output file: %w", err)
		}

		output.PrintInfo(fmt.Sprintf("Downloaded %d bytes to %s", n, injectDownloadOutput))
		return nil
	},
}

// ---------- inject metadata ----------

var injectMetadataCmd = &cobra.Command{
	Use:   "metadata",
	Short: "Show injection metadata (fault types, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		var resp client.APIResponse[any]
		if err := c.Get("/api/v2/injections/metadata", &resp); err != nil {
			return err
		}
		output.PrintJSON(resp.Data)
		return nil
	},
}

// ---------- init ----------

func init() {
	injectSubmitCmd.Flags().StringVar(&injectSubmitSpec, "spec", "", "Path to injection spec YAML file")

	injectListCmd.Flags().StringVar(&injectListState, "state", "", "Filter by state")
	injectListCmd.Flags().StringVar(&injectListFaultType, "fault-type", "", "Filter by fault type")
	injectListCmd.Flags().StringVar(&injectListLabels, "labels", "", "Filter by labels (key=val,...)")
	injectListCmd.Flags().IntVar(&injectListPage, "page", 1, "Page number")
	injectListCmd.Flags().IntVar(&injectListSize, "size", 20, "Page size")

	injectSearchCmd.Flags().StringVar(&injectSearchNamePattern, "name-pattern", "", "Name pattern to search for")
	injectSearchCmd.Flags().StringVar(&injectSearchLabels, "labels", "", "Labels to filter (key=val,...)")

	injectDownloadCmd.Flags().StringVarP(&injectDownloadOutput, "output-file", "O", "", "Output file path")

	injectCmd.AddCommand(injectSubmitCmd)
	injectCmd.AddCommand(injectListCmd)
	injectCmd.AddCommand(injectGetCmd)
	injectCmd.AddCommand(injectSearchCmd)
	injectCmd.AddCommand(injectLogsCmd)
	injectCmd.AddCommand(injectFilesCmd)
	injectCmd.AddCommand(injectDownloadCmd)
	injectCmd.AddCommand(injectMetadataCmd)
}
