package container

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestBuildGatewayBuildImageRef(t *testing.T) {
	gateway := &BuildGateway{
		registry:  "registry.example.com",
		namespace: "team-a",
	}

	if got := gateway.BuildImageRef("demo", "v1"); got != "registry.example.com/team-a/demo:v1" {
		t.Fatalf("unexpected image ref: %s", got)
	}
}

func TestBuildGatewayPrepareGitHubSourceCopiesSubPath(t *testing.T) {
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(filepath.Join(repoDir, "subdir"), 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, "subdir", "payload.txt"), []byte("payload"), 0o644); err != nil {
		t.Fatalf("write payload: %v", err)
	}

	runGit(t, repoDir, "init")
	runGit(t, repoDir, "config", "user.email", "codex@example.com")
	runGit(t, repoDir, "config", "user.name", "Codex")
	runGit(t, repoDir, "add", ".")
	runGit(t, repoDir, "commit", "-m", "init")

	viper.Set("jfs.container_path", tmpDir)
	gateway := &BuildGateway{
		containerBasePath: tmpDir,
		registry:          "registry.example.com",
		namespace:         "team-a",
		repoURLBuilder: func(*SubmitBuildContainerReq) string {
			return repoDir
		},
		commandRunner: exec.Command,
	}

	req := &SubmitBuildContainerReq{
		ImageName:        "demo",
		GithubRepository: "owner/repo",
		GithubBranch:     "master",
		SubPath:          "subdir",
	}

	targetDir, err := gateway.PrepareGitHubSource(req)
	if err != nil {
		t.Fatalf("PrepareGitHubSource failed: %v", err)
	}

	if filepath.Base(targetDir) == "subdir" {
		t.Fatalf("expected copied final directory, got raw subdir path: %s", targetDir)
	}

	content, err := os.ReadFile(filepath.Join(targetDir, "payload.txt"))
	if err != nil {
		t.Fatalf("read copied file: %v", err)
	}
	if string(content) != "payload" {
		t.Fatalf("unexpected copied content: %s", string(content))
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(output))
	}
}
