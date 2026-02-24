package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/sync/errgroup"
)

func statusBadge(status string) string {
	switch status {
	case "success":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✓ " + status)
	case "failure":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("✗ " + status)
	case "in_progress":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("⟳ " + status)
	default:
		return status
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: gh-util [flags]\n\nFlags:\n  -p, --pipeline   Open Azure DevOps pipeline runs for the current branch\n  -pr              Open pull requests for the current branch\n")
}

func runPipeline(info RepoInfo) error {
	pipelines, err := fetchPipelines(info.Owner, info.Repo, info.Branch)
	if err != nil {
		return err
	}

	if len(pipelines) == 0 {
		fmt.Printf("No Azure Pipelines check runs found on branch '%s'.\n", info.Branch)
		fmt.Println("This branch may not have triggered a pipeline yet. Try pushing first.")
		return nil
	}

	if len(pipelines) == 1 {
		fmt.Printf("Opening: %s [%s]\n", pipelines[0].Name, pipelines[0].Status)
		return openURL(pipelines[0].URL)
	}

	options := make([]huh.Option[int], len(pipelines))
	for i, p := range pipelines {
		label := fmt.Sprintf("%-55s %s", p.Name, statusBadge(p.Status))
		options[i] = huh.NewOption(label, i)
	}

	var selected int
	err = huh.NewSelect[int]().
		Title(fmt.Sprintf("Pipelines — %s/%s (%s)", info.Owner, info.Repo, info.Branch)).
		Options(options...).
		Value(&selected).
		Run()
	if err != nil {
		return err
	}

	return openURL(pipelines[selected].URL)
}

func prDirection(pr PR, branch string) string {
	if pr.HeadRef == branch {
		return fmt.Sprintf("%s → %s", pr.HeadRef, pr.BaseRef)
	}
	return fmt.Sprintf("%s → %s", pr.HeadRef, pr.BaseRef)
}

func runPR(info RepoInfo) error {
	var prsFrom, prsInto []PR

	g := new(errgroup.Group)
	g.Go(func() error {
		var err error
		prsFrom, err = fetchPRs(info.Branch)
		return err
	})
	g.Go(func() error {
		var err error
		prsInto, err = fetchPRsInto(info.Branch)
		return err
	})
	if err := g.Wait(); err != nil {
		return err
	}

	// Deduplicate: a PR from this branch could also target this branch (unlikely but possible)
	seen := make(map[int]bool)
	var prs []PR
	for _, pr := range prsFrom {
		seen[pr.Number] = true
		prs = append(prs, pr)
	}
	for _, pr := range prsInto {
		if !seen[pr.Number] {
			prs = append(prs, pr)
		}
	}

	if len(prs) == 0 {
		fmt.Printf("No open pull requests found for branch '%s'.\n", info.Branch)
		return nil
	}

	if len(prs) == 1 {
		pr := prs[0]
		draft := ""
		if pr.Draft {
			draft = " (draft)"
		}
		fmt.Printf("Opening: #%d %s [%s]%s\n", pr.Number, pr.Title, prDirection(pr, info.Branch), draft)
		return openURL(pr.URL)
	}

	options := make([]huh.Option[int], len(prs))
	for i, pr := range prs {
		draft := ""
		if pr.Draft {
			draft = " [draft]"
		}
		dir := prDirection(pr, info.Branch)
		label := fmt.Sprintf("#%-6d %s (%s)%s", pr.Number, pr.Title, dir, draft)
		options[i] = huh.NewOption(label, i)
	}

	var selected int
	if err := huh.NewSelect[int]().
		Title(fmt.Sprintf("Pull Requests — %s", info.Branch)).
		Options(options...).
		Value(&selected).
		Run(); err != nil {
		return err
	}

	return openURL(prs[selected].URL)
}

func main() {
	pipelineFlag := flag.Bool("p", false, "Open Azure DevOps pipeline runs")
	prFlag := flag.Bool("pr", false, "Open pull requests for the current branch")

	// Support --pipeline as alias for -p
	flag.BoolVar(pipelineFlag, "pipeline", false, "Open Azure DevOps pipeline runs")

	flag.Usage = usage
	flag.Parse()

	if !*pipelineFlag && !*prFlag {
		usage()
		os.Exit(1)
	}

	info, err := getRepoInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *pipelineFlag {
		err = runPipeline(info)
	} else {
		err = runPR(info)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
