package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

func IsValidDockerTag(tag string) error {
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}

	if len(tag) > 128 {
		return fmt.Errorf("tag too long (max 128 characters)")
	}

	tagPattern := `^[a-zA-Z0-9_][a-zA-Z0-9_.-]*$`
	matched, err := regexp.MatchString(tagPattern, tag)
	if err != nil {
		return fmt.Errorf("pattern match error: %v", err)
	}
	if !matched {
		return fmt.Errorf("tag contains invalid characters")
	}

	if strings.HasPrefix(tag, ".") || strings.HasPrefix(tag, "-") {
		return fmt.Errorf("tag cannot start with '.' or '-'")
	}

	return nil
}

func IsValidGitHubBranch(branch string) error {
	if len(branch) == 0 || len(branch) > 250 {
		return fmt.Errorf("branch name must be between 1 and 250 characters")
	}

	if strings.HasPrefix(branch, "/") || strings.HasSuffix(branch, "/") {
		return fmt.Errorf("branch name cannot start or end with '/'")
	}

	if strings.Contains(branch, "..") {
		return fmt.Errorf("branch name cannot contain '..'")
	}

	branchPattern := `^[a-zA-Z0-9._/-]+$`
	if matched, _ := regexp.MatchString(branchPattern, branch); !matched {
		return fmt.Errorf("branch name contains invalid characters")
	}

	return nil
}

func IsValidGitHubCommit(commit string) error {
	commitPattern := `^[a-fA-F0-9]{7,64}$`
	if matched, _ := regexp.MatchString(commitPattern, commit); !matched {
		return fmt.Errorf("invalid commit hash format")
	}

	return nil
}

func IsValidGitHubPath(path string) error {
	if strings.Contains(path, "..") {
		return fmt.Errorf("path cannot contain '..'")
	}

	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("path cannot start with '/'")
	}

	return nil
}

func IsValidGitHubToken(token string) error {
	tokenPrefixes := []string{"ghp_", "gho_", "ghu_", "ghs_", "ghr_"}
	hasValidPrefix := false

	for _, prefix := range tokenPrefixes {
		if strings.HasPrefix(token, prefix) {
			hasValidPrefix = true
			break
		}
	}

	if !hasValidPrefix {
		return fmt.Errorf("token does not have a valid GitHub token prefix")
	}

	if len(token) < 40 || len(token) > 255 {
		return fmt.Errorf("token length is invalid")
	}

	return nil
}

func IsValidEnvVar(envVar string) error {
	if envVar == "" {
		return fmt.Errorf("environment variable cannot be empty")
	}

	if len(envVar) > 128 {
		return fmt.Errorf("environment variable name too long (max 128 characters)")
	}

	envVarPattern := `^[A-Z_][A-Z0-9_]*$`
	matched, err := regexp.MatchString(envVarPattern, envVar)
	if err != nil {
		return fmt.Errorf("pattern match error: %v", err)
	}
	if !matched {
		return fmt.Errorf("environment variable must contain only uppercase letters, numbers, and underscores, and start with a letter or underscore")
	}

	return nil
}

func IsValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func ToSnakeCase(s string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
