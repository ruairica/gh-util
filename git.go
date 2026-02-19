package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type RepoInfo struct {
	Owner  string
	Repo   string
	Branch string
}

func getRepoInfo() (RepoInfo, error) {
	// Get remote URL
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return RepoInfo{}, fmt.Errorf("not in a git repository or no 'origin' remote found")
	}
	remoteURL := strings.TrimSpace(string(out))

	// Parse owner/repo from HTTPS or SSH URL
	re := regexp.MustCompile(`github\.com[:/]([^/]+)/([^/.]+?)(?:\.git)?$`)
	matches := re.FindStringSubmatch(remoteURL)
	if matches == nil {
		return RepoInfo{}, fmt.Errorf("remote is not a GitHub URL: %s", remoteURL)
	}
	owner := matches[1]
	repo := matches[2]

	// Get current branch
	branchOut, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	branch := "main"
	if err == nil {
		branch = strings.TrimSpace(string(branchOut))
	}

	return RepoInfo{Owner: owner, Repo: repo, Branch: branch}, nil
}
