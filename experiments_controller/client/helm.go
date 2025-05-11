package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/CUHK-SE-Group/rcabench/tracing"
	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/yaml"
)

// HelmClient represents a client for interacting with Helm
type HelmClient struct {
	namespace    string
	actionConfig *action.Configuration
	settings     *cli.EnvSettings
}

// NewHelmClient creates a new Helm client with the specified namespace
func NewHelmClient(namespace string) (*HelmClient, error) {
	settings := cli.New()
	settings.SetNamespace(namespace)

	actionConfig := new(action.Configuration)
	configFlags := genericclioptions.NewConfigFlags(true)
	configFlags.Namespace = &namespace

	if err := actionConfig.Init(configFlags, namespace, os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		return nil, fmt.Errorf("failed to initialize Helm action configuration: %w", err)
	}

	return &HelmClient{
		namespace:    namespace,
		actionConfig: actionConfig,
		settings:     settings,
	}, nil
}

// AddRepo adds a Helm repository with the given name and URL
func (c *HelmClient) AddRepo(name, url string) error {
	repoFile := c.settings.RepositoryConfig

	// Ensure the repository directory exists
	err := os.MkdirAll(c.settings.RepositoryCache, 0755)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("could not create repository cache directory: %w", err)
	}

	// Check if repo file exists
	b, err := os.ReadFile(repoFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not read repository file: %w", err)
	}

	var f repo.File
	if err == nil {
		if err := yaml.Unmarshal(b, &f); err != nil {
			return fmt.Errorf("cannot unmarshal repository file: %w", err)
		}
	}

	// Check if the repo already exists
	if f.Has(name) {
		// Repository already exists, nothing to do
		return nil
	}

	// Create new repository entry
	entry := &repo.Entry{
		Name: name,
		URL:  url,
	}

	r, err := repo.NewChartRepository(entry, getter.All(c.settings))
	if err != nil {
		return fmt.Errorf("failed to create chart repository: %w", err)
	}

	if _, err := r.DownloadIndexFile(); err != nil {
		return fmt.Errorf("looks like %q is not a valid chart repository or cannot be reached: %w", url, err)
	}

	f.Update(entry)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	return nil
}

// UpdateRepo updates all Helm repositories
func (c *HelmClient) UpdateRepo() error {
	repoFile := c.settings.RepositoryConfig

	// Read repo file
	b, err := os.ReadFile(repoFile)
	if err != nil {
		return fmt.Errorf("could not read repository file: %w", err)
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return fmt.Errorf("cannot unmarshal repository file: %w", err)
	}

	// Update each repository
	for _, entry := range f.Repositories {
		r, err := repo.NewChartRepository(entry, getter.All(c.settings))
		if err != nil {
			return fmt.Errorf("failed to create chart repository for %s: %w", entry.Name, err)
		}

		if _, err := r.DownloadIndexFile(); err != nil {
			return fmt.Errorf("failed to update repository %s: %w", entry.Name, err)
		}
	}

	return nil
}

func (c *HelmClient) SearchRepo(repoName string) ([]*repo.Entry, error) {
	repoFile := c.settings.RepositoryConfig

	b, err := os.ReadFile(repoFile)
	if err != nil {
		return nil, fmt.Errorf("could not read repository file: %w", err)
	}

	var f repo.File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("cannot unmarshal repository file: %w", err)
	}

	var repos []*repo.Entry
	for _, r := range f.Repositories {
		if repoName == "" || r.Name == repoName {
			repos = append(repos, r)
		}
	}

	return repos, nil
}

func (c *HelmClient) IsReleaseInstalled(releaseName string) (bool, error) {
	client := action.NewStatus(c.actionConfig)

	_, err := client.Run(releaseName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("failed to get release status: %w", err)
	}

	return true, nil
}

func (c *HelmClient) UninstallRelease(releaseName string) error {
	client := action.NewUninstall(c.actionConfig)
	client.Wait = true
	client.Timeout = 300 * time.Second

	_, err := client.Run(releaseName)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "release: not found") {
			logrus.Infof("Release %s is not installed, nothing to uninstall", releaseName)
			return nil
		}
		return fmt.Errorf("failed to uninstall release %s: %w", releaseName, err)
	}

	return nil
}

func (c *HelmClient) InstallRelease(ctx context.Context, releaseName, chartName string, vals map[string]interface{}) error {
	return tracing.WithSpan(ctx, func(ctx context.Context) error {
		now := time.Now()

		defer func() {
			log.Printf("InstallRelease took %s", time.Since(now))
		}()

		client := action.NewInstall(c.actionConfig)
		client.ReleaseName = releaseName
		client.Namespace = c.namespace
		client.Wait = true
		client.Timeout = 500 * time.Second
		client.CreateNamespace = true

		cp, err := client.ChartPathOptions.LocateChart(chartName, c.settings)
		if err != nil {
			return fmt.Errorf("failed to locate chart %s: %w", chartName, err)
		}

		chart, err := loader.Load(cp)
		if err != nil {
			return fmt.Errorf("failed to load chart %s: %w", chartName, err)
		}

		_, err = client.Run(chart, vals)
		if err != nil {
			return fmt.Errorf("failed to install release %s: %w", releaseName, err)
		}

		return nil
	})
}

func (c *HelmClient) InstallTrainTicket(ctx context.Context, releaseName, imageTag, nodePort string) error {
	installed, err := c.IsReleaseInstalled(releaseName)
	if err != nil {
		return err
	}

	// If installed, uninstall it first
	if installed {
		log.Printf("Uninstalling existing %s release", releaseName)
		if err := c.UninstallRelease(releaseName); err != nil {
			return err
		}
		// Wait a bit for resources to be cleaned up
		time.Sleep(5 * time.Second)
	} else {
		log.Printf("No existing %s release found", releaseName)
	}

	values := map[string]interface{}{
		"global": map[string]interface{}{
			"image": map[string]interface{}{
				"tag": imageTag,
			},
		},
		"services": map[string]interface{}{
			"tsUiDashboard": map[string]interface{}{
				"nodePort": nodePort,
			},
		},
	}

	return c.InstallRelease(ctx, releaseName, "train-ticket/trainticket", values)
}
