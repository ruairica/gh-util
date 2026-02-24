package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

type Pipeline struct {
	Name   string
	URL    string
	Status string
}

type PR struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	URL     string `json:"url"`
	State   string `json:"state"`
	Draft   bool   `json:"isDraft"`
	HeadRef string `json:"headRefName"`
	BaseRef string `json:"baseRefName"`
}

type checkRunsResponse struct {
	CheckRuns []checkRun `json:"check_runs"`
}

type checkRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	StartedAt  string `json:"started_at"`
	DetailsURL string `json:"details_url"`
	App        struct {
		Slug string `json:"slug"`
	} `json:"app"`
}

func fetchPipelines(owner, repo, branch string) ([]Pipeline, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/commits/%s/check-runs?per_page=100", owner, repo, branch)
	cmd := exec.Command("gh", "api", endpoint)
	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if exitErr, ok := err.(*exec.ExitError); ok && len(exitErr.Stderr) > 0 {
			stderr = "\n" + strings.TrimSpace(string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to call GitHub API%s\nEndpoint: %s\nIs 'gh' installed and authenticated? Run: gh auth login", stderr, endpoint)
	}

	var resp checkRunsResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	if len(resp.CheckRuns) == 0 {
		return nil, nil
	}

	// Deduplicate by name — keep the most recent run per check
	grouped := make(map[string][]checkRun)
	for _, cr := range resp.CheckRuns {
		grouped[cr.Name] = append(grouped[cr.Name], cr)
	}

	pipelines := make([]Pipeline, 0, len(grouped))
	for name, runs := range grouped {
		// Sort by started_at descending, pick latest
		sort.Slice(runs, func(i, j int) bool {
			return runs[i].StartedAt > runs[j].StartedAt
		})
		latest := runs[0]
		status := latest.Status
		if latest.Conclusion != "" {
			status = latest.Conclusion
		}
		pipelines = append(pipelines, Pipeline{
			Name:   name,
			URL:    latest.DetailsURL,
			Status: status,
		})
	}

	sort.Slice(pipelines, func(i, j int) bool {
		return strings.ToLower(pipelines[i].Name) < strings.ToLower(pipelines[j].Name)
	})

	return pipelines, nil
}

func fetchPRs(branch string) ([]PR, error) {
	out, err := exec.Command("gh", "pr", "list",
		"--head", branch,
		"--state", "open",
		"--json", "number,title,url,state,isDraft,headRefName,baseRefName",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs — is 'gh' installed and authenticated?\nRun: gh auth login")
	}

	var prs []PR
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse PR list: %w", err)
	}

	return prs, nil
}

func fetchPRsInto(branch string) ([]PR, error) {
	out, err := exec.Command("gh", "pr", "list",
		"--base", branch,
		"--state", "open",
		"--json", "number,title,url,state,isDraft,headRefName,baseRefName",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs — is 'gh' installed and authenticated?\nRun: gh auth login")
	}

	var prs []PR
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse PR list: %w", err)
	}

	return prs, nil
}
