package cmd

import (
	"strings"
	"testing"
)

func TestParseImageRef(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want imageRefParts
	}{
		{"default registry, single namespace", "library/nginx:1.25", imageRefParts{Registry: "docker.io", Namespace: "library", Repository: "nginx", Tag: "1.25"}},
		{"default registry, no namespace", "nginx:1.25", imageRefParts{Registry: "docker.io", Namespace: "", Repository: "nginx", Tag: "1.25"}},
		{"explicit registry with nested namespace", "docker.io/foo/bar/baz:tag", imageRefParts{Registry: "docker.io", Namespace: "foo/bar", Repository: "baz", Tag: "tag"}},
		{"ghcr with nested namespace", "ghcr.io/org/team/app:v1.2.3", imageRefParts{Registry: "ghcr.io", Namespace: "org/team", Repository: "app", Tag: "v1.2.3"}},
		{"localhost registry with port", "localhost:5000/team/app:dev", imageRefParts{Registry: "localhost:5000", Namespace: "team", Repository: "app", Tag: "dev"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseImageRef(tc.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("parseImageRef(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseImageRef_Rejections(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		wantSubs string
	}{
		{"no tag", "docker.io/library/nginx", "missing ':<tag>'"},
		{"no tag bare", "nginx", "missing ':<tag>'"},
		{"empty tag", "nginx:", "empty tag"},
		{"digest ref", "nginx@sha256:abcdef", "digest"},
		{"empty", "   ", "empty"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseImageRef(tc.in)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tc.in)
			}
			if !strings.Contains(err.Error(), tc.wantSubs) {
				t.Fatalf("expected error %q to contain %q", err.Error(), tc.wantSubs)
			}
		})
	}
}
