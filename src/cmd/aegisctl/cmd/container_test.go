package cmd

import (
	"encoding/json"
	"testing"
)

func TestContainerDetailDeserialization(t *testing.T) {
	t.Run("two versions", func(t *testing.T) {
		raw := `{
			"id": 1,
			"name": "my-container",
			"type": "algorithm",
			"status": "ready",
			"versions": [
				{"id": 10, "name": "v1.0", "image_ref": "registry/img:v1.0", "usage": 3, "updated_at": "2026-01-01"},
				{"id": 11, "name": "v1.1", "image_ref": "registry/img:v1.1", "usage": 5, "updated_at": "2026-01-02"}
			],
			"created_at": "2025-12-01",
			"updated_at": "2026-01-02"
		}`

		var detail containerDetail
		if err := json.Unmarshal([]byte(raw), &detail); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if len(detail.Versions) != 2 {
			t.Fatalf("expected 2 versions, got %d", len(detail.Versions))
		}
		if detail.Versions[0].Name != "v1.0" {
			t.Errorf("expected first version name %q, got %q", "v1.0", detail.Versions[0].Name)
		}
	})

	t.Run("empty versions array", func(t *testing.T) {
		raw := `{"id": 2, "name": "ctr", "versions": []}`

		var detail containerDetail
		if err := json.Unmarshal([]byte(raw), &detail); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if len(detail.Versions) != 0 {
			t.Fatalf("expected 0 versions, got %d", len(detail.Versions))
		}
	})

	t.Run("no versions field", func(t *testing.T) {
		raw := `{"id": 3, "name": "ctr"}`

		var detail containerDetail
		if err := json.Unmarshal([]byte(raw), &detail); err != nil {
			t.Fatalf("unmarshal error: %v", err)
		}
		if detail.Versions != nil {
			t.Fatalf("expected nil Versions, got %v", detail.Versions)
		}
	})
}

func TestContainerGetOutputComputedFields(t *testing.T) {
	t.Run("with versions", func(t *testing.T) {
		out := containerGetOutput{
			containerDetail: containerDetail{
				ID:   1,
				Name: "my-container",
				Versions: []containerVersionItem{
					{ID: 10, Name: "v1.0"},
					{ID: 11, Name: "v1.1"},
				},
			},
			DefaultVersion: "v1.1",
			VersionCount:   2,
		}

		data, err := json.Marshal(out)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}

		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("unmarshal to map error: %v", err)
		}

		if v, ok := m["default_version"]; !ok {
			t.Fatal("missing default_version key")
		} else if v != "v1.1" {
			t.Errorf("expected default_version %q, got %v", "v1.1", v)
		}

		if v, ok := m["version_count"]; !ok {
			t.Fatal("missing version_count key")
		} else if v != float64(2) {
			t.Errorf("expected version_count 2, got %v", v)
		}
	})

	t.Run("zero versions", func(t *testing.T) {
		out := containerGetOutput{
			containerDetail: containerDetail{
				ID:   2,
				Name: "empty-ctr",
			},
			DefaultVersion: "(none)",
			VersionCount:   0,
		}

		data, err := json.Marshal(out)
		if err != nil {
			t.Fatalf("marshal error: %v", err)
		}

		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			t.Fatalf("unmarshal to map error: %v", err)
		}

		if v := m["default_version"]; v != "(none)" {
			t.Errorf("expected default_version %q, got %v", "(none)", v)
		}
		if v := m["version_count"]; v != float64(0) {
			t.Errorf("expected version_count 0, got %v", v)
		}
	})
}
