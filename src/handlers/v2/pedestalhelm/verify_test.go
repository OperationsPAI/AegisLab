package pedestalhelm

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakeRunner struct {
	addErr    error
	addOut    string
	updateErr error
	updateOut string
	pullErr   error
	pullOut   string

	addCalled    bool
	updateCalled bool
	pullCalled   bool
}

func (f *fakeRunner) RepoAdd(name, url string) (string, error) {
	f.addCalled = true
	return f.addOut, f.addErr
}

func (f *fakeRunner) RepoUpdate() (string, error) {
	f.updateCalled = true
	return f.updateOut, f.updateErr
}

func (f *fakeRunner) Pull(repo, chart, version, dest string) (string, error) {
	f.pullCalled = true
	return f.pullOut, f.pullErr
}

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "values.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp yaml: %v", err)
	}
	return p
}

func TestRun_AllGreen(t *testing.T) {
	cfg := Config{
		ChartName: "pedestal",
		Version:   "1.2.3",
		RepoURL:   "https://example.com/charts",
		RepoName:  "aegis",
		ValueFile: writeTempYAML(t, "image:\n  repository: nginx\n  tag: \"1.25\"\n"),
	}
	r := &fakeRunner{}
	got := Run(r, cfg, VerifyValueFile)
	if !got.OK {
		t.Fatalf("expected OK=true, got %+v", got)
	}
	if !r.addCalled || !r.updateCalled || !r.pullCalled {
		t.Fatalf("expected all helm calls, got %+v", r)
	}
	want := []string{"repo_add", "repo_update", "helm_pull", "value_file"}
	if len(got.Checks) != len(want) {
		t.Fatalf("expected %v checks, got %+v", want, got.Checks)
	}
	for i, n := range want {
		if got.Checks[i].Name != n {
			t.Fatalf("check[%d]=%q want %q", i, got.Checks[i].Name, n)
		}
	}
}

func TestRun_RepoAddFailsShortCircuits(t *testing.T) {
	cfg := Config{RepoName: "aegis", RepoURL: "https://bad"}
	r := &fakeRunner{addErr: errors.New("boom"), addOut: "stderr text"}
	got := Run(r, cfg, VerifyValueFile)
	if got.OK {
		t.Fatalf("expected OK=false")
	}
	if len(got.Checks) != 1 || got.Checks[0].Name != "repo_add" || got.Checks[0].OK {
		t.Fatalf("unexpected checks: %+v", got.Checks)
	}
	if !strings.Contains(got.Checks[0].Detail, "stderr text") {
		t.Fatalf("stderr should be surfaced, got %q", got.Checks[0].Detail)
	}
	if r.updateCalled || r.pullCalled {
		t.Fatalf("later steps must not run after repo_add failure")
	}
}

func TestRun_PullFailDoesNotSkipValueFile(t *testing.T) {
	cfg := Config{
		ChartName: "p", Version: "1", RepoURL: "u", RepoName: "n",
		ValueFile: writeTempYAML(t, "image:\n  repository: nginx\n  tag: \"v1\"\n"),
	}
	r := &fakeRunner{pullErr: errors.New("nope"), pullOut: "not found"}
	got := Run(r, cfg, VerifyValueFile)
	if got.OK {
		t.Fatal("expected OK=false")
	}
	var gotPull, gotVF bool
	for _, c := range got.Checks {
		if c.Name == "helm_pull" {
			gotPull = true
			if c.OK {
				t.Fatal("helm_pull should be false")
			}
			if !strings.Contains(c.Detail, "not found") {
				t.Fatalf("pull stderr not surfaced: %q", c.Detail)
			}
		}
		if c.Name == "value_file" {
			gotVF = true
			if !c.OK {
				t.Fatalf("value_file should still run and succeed: %+v", c)
			}
		}
	}
	if !gotPull || !gotVF {
		t.Fatalf("both helm_pull and value_file expected, got %+v", got.Checks)
	}
}

func TestRun_BadValueFile(t *testing.T) {
	bad := writeTempYAML(t, "image:\n  repository: [nope\n")
	cfg := Config{
		ChartName: "p", Version: "1", RepoURL: "u", RepoName: "n",
		ValueFile: bad,
	}
	r := &fakeRunner{}
	got := Run(r, cfg, VerifyValueFile)
	if got.OK {
		t.Fatal("expected OK=false for bad yaml")
	}
	last := got.Checks[len(got.Checks)-1]
	if last.Name != "value_file" || last.OK {
		t.Fatalf("expected value_file to fail, got %+v", last)
	}
}

func TestVerifyValueFile_RejectsNonStringRepository(t *testing.T) {
	p := writeTempYAML(t, "image:\n  repository:\n    - a\n    - b\n")
	if err := VerifyValueFile(p); err == nil {
		t.Fatal("expected error for non-string image.repository")
	}
}

func TestVerifyValueFile_AcceptsIntTag(t *testing.T) {
	p := writeTempYAML(t, "image:\n  repository: nginx\n  tag: 1\n")
	if err := VerifyValueFile(p); err != nil {
		t.Fatalf("int tag should be accepted, got %v", err)
	}
}

func TestVerifyValueFile_NoImageSection(t *testing.T) {
	p := writeTempYAML(t, "service:\n  type: ClusterIP\n")
	if err := VerifyValueFile(p); err != nil {
		t.Fatalf("absent image section should be fine, got %v", err)
	}
}
