// Package pedestal holds dry-run verification pipeline for helm_configs rows
// used by the /api/v2/pedestal/helm/:id/verify endpoint.
package pedestal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the minimal projection of model.HelmConfig that the verify pipeline needs.
type Config struct {
	ChartName string
	Version   string
	RepoURL   string
	RepoName  string
	ValueFile string
}

// Check is a single step outcome.
type Check struct {
	Name   string `json:"name"`
	OK     bool   `json:"ok"`
	Detail string `json:"detail,omitempty"`
}

// Result is the aggregated verification result.
type Result struct {
	OK     bool    `json:"ok"`
	Checks []Check `json:"checks"`
}

// Runner abstracts helm CLI invocations so tests can stub them out.
type Runner interface {
	RepoAdd(name, url string) (string, error)
	RepoUpdate() (string, error)
	Pull(repo, chart, version, destDir string) (string, error)
}

// RealRunner shells out to the real `helm` binary.
type RealRunner struct{}

func (RealRunner) RepoAdd(name, url string) (string, error) {
	out, err := exec.Command("helm", "repo", "add", name, url, "--force-update").CombinedOutput()
	return string(out), err
}

func (RealRunner) RepoUpdate() (string, error) {
	out, err := exec.Command("helm", "repo", "update").CombinedOutput()
	return string(out), err
}

func (RealRunner) Pull(repo, chart, version, destDir string) (string, error) {
	out, err := exec.Command("helm", "pull", fmt.Sprintf("%s/%s", repo, chart),
		"--version", version, "--destination", destDir).CombinedOutput()
	return string(out), err
}

// Run drives the check pipeline.
func Run(runner Runner, cfg Config, valueFileVerifier func(string) error) Result {
	checks := make([]Check, 0, 4)

	if out, err := runner.RepoAdd(cfg.RepoName, cfg.RepoURL); err != nil {
		checks = append(checks, Check{
			Name: "repo_add", OK: false,
			Detail: fmt.Sprintf("helm repo add failed: %v\n%s", err, out),
		})
		return Result{OK: false, Checks: checks}
	}
	checks = append(checks, Check{Name: "repo_add", OK: true})

	if out, err := runner.RepoUpdate(); err != nil {
		checks = append(checks, Check{
			Name: "repo_update", OK: false,
			Detail: fmt.Sprintf("helm repo update failed: %v\n%s", err, out),
		})
		return Result{OK: false, Checks: checks}
	}
	checks = append(checks, Check{Name: "repo_update", OK: true})

	tmpDir, err := os.MkdirTemp("", "aegis-helm-verify-")
	if err != nil {
		return Result{OK: false, Checks: append(checks, Check{
			Name: "helm_pull", OK: false,
			Detail: "could not create tmp dir: " + err.Error(),
		})}
	}
	defer os.RemoveAll(tmpDir)

	allOK := true
	if out, err := runner.Pull(cfg.RepoName, cfg.ChartName, cfg.Version, tmpDir); err != nil {
		allOK = false
		checks = append(checks, Check{
			Name: "helm_pull", OK: false,
			Detail: fmt.Sprintf("helm pull failed: %v\n%s", err, out),
		})
	} else {
		checks = append(checks, Check{Name: "helm_pull", OK: true})
	}

	if cfg.ValueFile != "" {
		if err := valueFileVerifier(cfg.ValueFile); err != nil {
			allOK = false
			checks = append(checks, Check{Name: "value_file", OK: false, Detail: err.Error()})
		} else {
			checks = append(checks, Check{Name: "value_file", OK: true})
		}
	}

	return Result{OK: allOK, Checks: checks}
}

// VerifyValueFile opens the values file and asserts it parses as YAML.
func VerifyValueFile(path string) error {
	abs := path
	if !filepath.IsAbs(abs) {
		abs, _ = filepath.Abs(abs)
	}
	f, err := os.Open(abs)
	if err != nil {
		return fmt.Errorf("open value file %q: %w", path, err)
	}
	defer f.Close()

	var parsed map[string]any
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&parsed); err != nil {
		return fmt.Errorf("parse value file %q: %w", path, err)
	}

	if img, ok := parsed["image"].(map[string]any); ok {
		if repo, present := img["repository"]; present {
			if _, ok := repo.(string); !ok {
				return fmt.Errorf("image.repository is not a string in %q", path)
			}
		}
		if tag, present := img["tag"]; present {
			switch tag.(type) {
			case string, int, int64, float64:
			default:
				return fmt.Errorf("image.tag is not a scalar in %q", path)
			}
		}
	}
	return nil
}
