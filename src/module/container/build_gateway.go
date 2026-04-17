package containermodule

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"aegis/config"
	"aegis/utils"

	"github.com/sirupsen/logrus"
)

type BuildGateway struct {
	containerBasePath string
	registry          string
	namespace         string
	repoURLBuilder    func(*SubmitBuildContainerReq) string
	commandRunner     func(string, ...string) *exec.Cmd
}

func NewBuildGateway() *BuildGateway {
	return &BuildGateway{
		containerBasePath: config.GetString("jfs.container_path"),
		registry:          config.GetString("harbor.registry"),
		namespace:         config.GetString("harbor.namespace"),
		repoURLBuilder: func(req *SubmitBuildContainerReq) string {
			repoURL := fmt.Sprintf("https://github.com/%s.git", req.GithubRepository)
			if req.GithubToken != "" {
				repoURL = fmt.Sprintf("https://%s@github.com/%s.git", req.GithubToken, req.GithubRepository)
			}
			return repoURL
		},
		commandRunner: exec.Command,
	}
}

func (g *BuildGateway) BuildImageRef(imageName, tag string) string {
	return fmt.Sprintf("%s/%s/%s:%s", g.registry, g.namespace, imageName, tag)
}

func (g *BuildGateway) PrepareGitHubSource(req *SubmitBuildContainerReq) (string, error) {
	targetDir := filepath.Join(g.containerBasePath, req.ImageName, fmt.Sprintf("build_%d", time.Now().Unix()))
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create target directory: %w", err)
	}

	repoURL := g.repoURLBuilder(req)

	gitCmd := []string{"git", "clone"}
	if req.GithubBranch != "" {
		gitCmd = append(gitCmd, "--branch", req.GithubBranch, "--single-branch")
	}
	gitCmd = append(gitCmd, repoURL, targetDir)

	cmd := g.commandRunner(gitCmd[0], gitCmd[1:]...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	if req.GithubCommit != "" {
		cmd = g.commandRunner("git", "-C", targetDir, "checkout", req.GithubCommit)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to checkout commit %s: %w", req.GithubCommit, err)
		}
	}

	if req.SubPath != "" && req.SubPath != "." {
		sourcePath := filepath.Join(targetDir, req.SubPath)
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			return "", fmt.Errorf("sub path '%s' does not exist in repository", req.SubPath)
		}

		newTargetDir := filepath.Join(g.containerBasePath, req.ImageName, fmt.Sprintf("build_final_%d", time.Now().Unix()))
		if err := utils.CopyDir(sourcePath, newTargetDir); err != nil {
			return "", fmt.Errorf("failed to copy subdirectory: %w", err)
		}
		if err := os.RemoveAll(targetDir); err != nil {
			logrus.WithField("target_dir", targetDir).Warnf("failed to remove temporary directory: %v", err)
		}
		targetDir = newTargetDir
	}

	return targetDir, nil
}
