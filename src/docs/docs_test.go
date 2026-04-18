package docs_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestGeneratedAPIDocsContainCorePaths(t *testing.T) {
	baseDir := "."

	checkJSONContains(t, filepath.Join(baseDir, "openapi2", "swagger.json"), []string{
		`"/api/v2/auth/login"`,
		`"/api/v2/users"`,
		`"/api/v2/projects"`,
	})
	checkJSONContains(t, filepath.Join(baseDir, "openapi3", "openapi.json"), []string{
		`"/api/v2/auth/login"`,
		`"/api/v2/projects"`,
		`"/api/v2/users"`,
	})
	checkJSONContains(t, filepath.Join(baseDir, "converted", "sdk.json"), []string{
		`"/api/v2/auth/api-key/token"`,
		`"/api/v2/sdk/evaluations"`,
	})
	checkJSONContains(t, filepath.Join(baseDir, "converted", "runtime.json"), []string{
		`"/api/v2/executions/{execution_id}/detector_results"`,
		`"/api/v2/executions/{execution_id}/granularity_results"`,
	})
}

func TestAudienceFilteredDocsMatchOpenAPI3Extensions(t *testing.T) {
	openapi := readJSON(t, filepath.Join(".", "openapi3", "openapi.json"))

	checkAudienceMatches(t, openapi, filepath.Join(".", "converted", "sdk.json"), "sdk")
	checkAudienceMatches(t, openapi, filepath.Join(".", "converted", "runtime.json"), "runtime")
	checkAudienceMatches(t, openapi, filepath.Join(".", "converted", "portal.json"), "portal")
	checkAudienceMatches(t, openapi, filepath.Join(".", "converted", "admin.json"), "admin")
}

func checkJSONContains(t *testing.T, path string, fragments []string) {
	t.Helper()

	data := readJSONBytes(t, path)
	text := string(data)
	for _, fragment := range fragments {
		if !strings.Contains(text, fragment) {
			t.Fatalf("expected %s to contain %s", path, fragment)
		}
	}
}

func checkAudienceMatches(t *testing.T, openapi map[string]any, filteredPath string, audience string) {
	t.Helper()
	checkAudienceMatchesAny(t, openapi, filteredPath, []string{audience})
}

func checkAudienceMatchesAny(t *testing.T, openapi map[string]any, filteredPath string, audiences []string) {
	t.Helper()

	filtered := readJSON(t, filteredPath)
	want := collectAudienceOperations(t, openapi, audiences)
	got := collectOperationsFromDoc(t, filtered)

	if len(want) != len(got) {
		t.Fatalf("expected %s to have %d operations, got %d", filteredPath, len(want), len(got))
	}
	if strings.Join(want, "\n") != strings.Join(got, "\n") {
		t.Fatalf("unexpected operations in %s\nwant:\n%s\n\ngot:\n%s", filteredPath, strings.Join(want, "\n"), strings.Join(got, "\n"))
	}
}

func collectAudienceOperations(t *testing.T, doc map[string]any, audiences []string) []string {
	t.Helper()

	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		t.Fatalf("paths is not an object")
	}

	audienceSet := make(map[string]struct{}, len(audiences))
	for _, audience := range audiences {
		audienceSet[audience] = struct{}{}
	}

	var operations []string
	for path, opsValue := range paths {
		ops, ok := opsValue.(map[string]any)
		if !ok {
			continue
		}
		for method, specValue := range ops {
			spec, ok := specValue.(map[string]any)
			if !ok {
				continue
			}
			xAPIType, _ := spec["x-api-type"].(map[string]any)
			for audience := range audienceSet {
				if isAudienceEnabled(xAPIType, audience) {
					operations = append(operations, strings.ToUpper(method)+" "+path)
					break
				}
			}
		}
	}

	sort.Strings(operations)
	return operations
}

func isAudienceEnabled(xAPIType map[string]any, audience string) bool {
	value, ok := xAPIType[audience]
	if !ok {
		return false
	}

	switch typed := value.(type) {
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	case bool:
		return typed
	default:
		return false
	}
}

func collectOperationsFromDoc(t *testing.T, doc map[string]any) []string {
	t.Helper()

	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		t.Fatalf("paths is not an object")
	}

	var operations []string
	for path, opsValue := range paths {
		ops, ok := opsValue.(map[string]any)
		if !ok {
			continue
		}
		for method := range ops {
			operations = append(operations, strings.ToUpper(method)+" "+path)
		}
	}

	sort.Strings(operations)
	return operations
}

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()

	data := readJSONBytes(t, path)
	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return body
}

func readJSONBytes(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}
