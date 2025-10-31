package utils

import (
	"fmt"
	"regexp"
	"strings"
)

func IsValidGitHubLink(link string) error {
	if link == "" {
		return nil
	}

	parts := strings.Split(link, "/")
	numParts := len(parts)

	if numParts == 2 {
		return IsValidGitHubRepository(link)
	}

	if numParts >= 5 {
		repo := strings.Join(parts[0:2], "/")
		if err := IsValidGitHubRepository(repo); err != nil {
			return fmt.Errorf("invalid repository part in link: %w", err)
		}

		branchOrCommit := parts[3]
		if err := IsValidGitHubBranch(branchOrCommit); err != nil {
			if err := IsValidGitHubCommit(branchOrCommit); err != nil {
				return fmt.Errorf("invalid branch or commit hash in link: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("github link has an invalid number of segments")
}

func IsValidGitHubRepository(repo string) error {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return fmt.Errorf("repository must be in the format 'owner/repo'")
	}

	owner := parts[0]

	ownerPattern := `^[a-zA-Z0-9-]+$`
	repoPattern := `^[a-zA-Z0-9._-]{1,99}[a-zA-Z0-9_-]$|^[a-zA-Z0-9._-]{1,100}$`

	if matched, err := regexp.MatchString(ownerPattern, parts[0]); err != nil {
		return fmt.Errorf("internal regex error: %v", err)
	} else if !matched {
		return fmt.Errorf("invalid repository owner format")
	}

	if len(owner) < 1 || len(owner) > 39 {
		return fmt.Errorf("invalid repository owner format (length)")
	}

	if owner[0] == '-' || owner[len(owner)-1] == '-' {
		return fmt.Errorf("invalid repository owner format (starts/ends with hyphen)")
	}

	if strings.Contains(owner, "--") {
		return fmt.Errorf("invalid repository owner format (consecutive hyphens)")
	}

	if matched, err := regexp.MatchString(repoPattern, parts[1]); err != nil {
		return fmt.Errorf("internal regex error: %v", err)
	} else if !matched {
		return fmt.Errorf("invalid repository name format")
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
	if strings.Contains(branch, "@") {
		return fmt.Errorf("branch name cannot contain '@'")
	}

	branchPattern := `^[a-zA-Z0-9._/-]+$`
	if matched, err := regexp.MatchString(branchPattern, branch); err != nil {
		return fmt.Errorf("internal regex error: %v", err)
	} else if !matched {
		return fmt.Errorf("branch name contains invalid characters or sequences")
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

func IsValidGitHubToken(token string) error {
	tokenPrefixes := []string{"ghp_", "gho_", "ghu_", "ghs_", "ghr_"}

	prefixMatch := ""
	for _, prefix := range tokenPrefixes {
		if strings.HasPrefix(token, prefix) {
			prefixMatch = prefix
			break
		}
	}

	if prefixMatch == "" {
		return fmt.Errorf("token does not have a valid GitHub token prefix")
	}

	if prefixMatch == "ghp_" && len(token) != 40 {
		return fmt.Errorf("personal access token (ghp_) must be exactly 40 characters long")
	}

	suffix := token[len(prefixMatch):]
	if matched, err := regexp.MatchString(`^[a-fA-F0-9]+$`, suffix); err != nil {
		return fmt.Errorf("internal regex error: %v", err)
	} else if !matched {
		return fmt.Errorf("token contains invalid characters")
	}

	return nil
}
