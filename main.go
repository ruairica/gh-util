package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
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

func run() error {
	info, err := getRepoInfo()
	if err != nil {
		return err
	}

	pipelines, err := fetchPipelines(info.Owner, info.Repo, info.Branch)
	if err != nil {
		return err
	}

	if len(pipelines) == 0 {
		fmt.Printf("No Azure Pipelines check runs found on branch '%s'.\n", info.Branch)
		fmt.Println("This branch may not have triggered a pipeline yet. Try pushing first.")
		return nil
	}

	// Single pipeline — open directly
	if len(pipelines) == 1 {
		fmt.Printf("Opening: %s [%s]\n", pipelines[0].Name, pipelines[0].Status)
		return openURL(pipelines[0].URL)
	}

	// Multiple pipelines — interactive select
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

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
