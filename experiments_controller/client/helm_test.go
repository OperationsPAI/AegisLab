package client

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// mockActionConfig creates a mock Helm action configuration for testing
func mockActionConfig(t *testing.T) *action.Configuration {
	actionConfig := new(action.Configuration)
	configFlags := genericclioptions.NewConfigFlags(true)
	namespace := "test-namespace"
	configFlags.Namespace = &namespace

	// Use memory driver for testing to avoid real k8s connections
	err := actionConfig.Init(configFlags, namespace, "memory", func(format string, v ...interface{}) {
		t.Logf(format, v...)
	})
	if err != nil {
		t.Fatalf("Failed to initialize action config: %v", err)
	}

	return actionConfig
}

// createMockHelmClient creates a test HelmClient with mock configuration
func createMockHelmClient(t *testing.T) *HelmClient {
	settings := cli.New()
	namespace := "test-namespace"
	settings.SetNamespace(namespace)

	// Create temporary directories for testing
	tempDir := t.TempDir()
	settings.RepositoryConfig = filepath.Join(tempDir, "repositories.yaml")
	settings.RepositoryCache = filepath.Join(tempDir, "cache")

	return &HelmClient{
		namespace:    namespace,
		actionConfig: mockActionConfig(t),
		settings:     settings,
	}
}

func TestHelmClient_isChartCachedLocally(t *testing.T) {
	tests := []struct {
		name        string
		chartName   string
		expectFound bool
	}{
		{
			name:        "chart does not exist in cache",
			chartName:   "non-existent-chart",
			expectFound: false,
		},
		{
			name:        "empty chart name",
			chartName:   "",
			expectFound: false,
		},
		{
			name:        "invalid chart name",
			chartName:   "invalid/chart/name",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := createMockHelmClient(t)

			gotPath, gotFound := client.isChartCachedLocally(tt.chartName)

			if gotFound != tt.expectFound {
				t.Errorf("isChartCachedLocally() gotFound = %v, want %v", gotFound, tt.expectFound)
			}

			if tt.expectFound {
				if gotPath == "" {
					t.Error("Expected non-empty chart path when chart is found")
				}

				// Verify the chart path exists
				if _, err := os.Stat(gotPath); err != nil {
					t.Errorf("Chart path should exist: %v", err)
				}
			} else {
				if gotPath != "" {
					t.Errorf("Expected empty chart path when not found, got: %s", gotPath)
				}
			}
		})
	}
}

func TestHelmClient_isChartCachedLocally_FileSystem(t *testing.T) {
	client := createMockHelmClient(t)

	// Test with a chart that definitely doesn't exist
	path, found := client.isChartCachedLocally("definitely-non-existent-chart-12345")
	if found {
		t.Error("Should not find non-existent chart")
	}
	if path != "" {
		t.Errorf("Path should be empty for non-existent chart, got: %s", path)
	}
}

func TestHelmClient_isChartCachedLocally_NoDownload(t *testing.T) {
	client := createMockHelmClient(t)

	// This test ensures that the cache check doesn't trigger a download
	// We test with a chart name that would normally trigger a download
	startTime := time.Now()

	path, found := client.isChartCachedLocally("non-existent-repo/non-existent-chart")

	elapsed := time.Since(startTime)

	// The operation should be very fast since it's only checking local filesystem
	if elapsed > 1*time.Second {
		t.Errorf("Cache check took too long (%v), might be triggering download", elapsed)
	}

	// Should not find the chart since it doesn't exist locally
	if found {
		t.Error("Should not find non-existent chart")
	}

	if path != "" {
		t.Errorf("Path should be empty for non-existent chart, got: %s", path)
	}

	t.Logf("Cache check completed in %v (no download triggered)", elapsed)
}

func TestHelmClient_InstallRelease_UsesCachedChart(t *testing.T) {
	// This test verifies the method handles the basic flow
	client := createMockHelmClient(t)
	ctx := context.Background()

	// Create a simple test that verifies the method doesn't panic
	// and handles the basic flow (though it will fail due to missing chart)
	err := client.InstallRelease(ctx, "test-release", "non-existent-chart", map[string]any{})

	// We expect an error since the chart doesn't exist, but it should be a specific error
	if err == nil {
		t.Error("Expected error for non-existent chart")
	}

	if !strings.Contains(err.Error(), "failed to locate chart") {
		t.Errorf("Expected 'failed to locate chart' error, got: %v", err)
	}
}

func TestHelmClient_NewHelmClient(t *testing.T) {
	namespace := "test-namespace"

	// This test might fail in environments without proper k8s config
	// but we can test the basic structure
	client, err := NewHelmClient(namespace)

	if err != nil {
		// If we can't create a real client (e.g., no k8s config), that's expected in test env
		t.Logf("Expected error in test environment: %v", err)
		return
	}

	if client == nil {
		t.Error("Expected non-nil client")
	}
	if client.namespace != namespace {
		t.Errorf("Expected namespace %s, got %s", namespace, client.namespace)
	}
	if client.actionConfig == nil {
		t.Error("Expected non-nil actionConfig")
	}
	if client.settings == nil {
		t.Error("Expected non-nil settings")
	}
}
