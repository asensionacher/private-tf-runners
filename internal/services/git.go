package services

import (
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/private-tf-runners/server/internal/models"
)

type GitService struct{}

func NewGitService() *GitService {
	return &GitService{}
}

func (s *GitService) FetchRepoInfo(gitURL string) (*models.RepoInfo, error) {
	branches, err := s.getBranches(gitURL)
	if err != nil {
		return nil, err
	}

	tags, err := s.getTags(gitURL)
	if err != nil {
		return nil, err
	}

	return &models.RepoInfo{
		Branches: branches,
		Tags:     tags,
	}, nil
}

func (s *GitService) getBranches(gitURL string) ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", gitURL)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	branches := []string{}
	seen := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		ref := parts[1]
		ref = strings.TrimPrefix(ref, "refs/heads/")
		if !seen[ref] {
			seen[ref] = true
			branches = append(branches, ref)
		}
	}
	return branches, nil
}

func (s *GitService) getTags(gitURL string) ([]string, error) {
	cmd := exec.Command("git", "ls-remote", "--tags", gitURL)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	tags := []string{}
	seen := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		ref := parts[1]
		ref = strings.TrimPrefix(ref, "refs/tags/")
		if isVersionTag(ref) && !seen[ref] {
			seen[ref] = true
			tags = append(tags, ref)
		}
	}
	return tags, nil
}

func isVersionTag(tag string) bool {
	tag = strings.TrimSuffix(tag, "^{}")
	versionRegex := regexp.MustCompile(`^v?\d+\.\d+`)
	return versionRegex.MatchString(tag)
}