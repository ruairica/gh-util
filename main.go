package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/sync/errgroup"
)

// checkNameWidth is the column width check names are padded to so status
// badges align; child rows subtract their "├ " prefix from it.
const checkNameWidth = 55

// splitSubStage splits a check name of the form "X (Stage Job)" into the
// parent name "X" and the stage text, when a check named exactly "X" exists
// in names. The longest matching parent wins, so "X (a) (b)" nests under
// "X (a)" when that check also exists.
func splitSubStage(name string, names map[string]bool) (parent, stage string, ok bool) {
	base, hadParen := strings.CutSuffix(name, ")")
	if !hadParen {
		return "", "", false
	}
	for idx := strings.LastIndex(base, " ("); idx > 0; idx = strings.LastIndex(base[:idx], " (") {
		if names[base[:idx]] && idx+2 < len(base) {
			return base[:idx], base[idx+2:], true
		}
	}
	return "", "", false
}

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
	fmt.Fprintf(os.Stderr, "Usage: gh-util [flags] [branch]\n\nFlags:\n  -ci     Open CI check runs (current branch, or [branch] if given)\n  -pr     Open pull requests (current branch, or [branch] if given)\n  -wait   Poll until check runs or pull requests are available (use with -ci or -pr)\n")
}

func runChecks(info RepoInfo, wait bool) error {
	var checks []Check
	for {
		var err error
		checks, err = fetchChecks(info.Owner, info.Repo, info.Branch)
		if err != nil {
			return err
		}
		if len(checks) > 0 {
			break
		}
		if !wait {
			fmt.Printf("No check runs found on branch '%s'.\n", info.Branch)
			fmt.Println("This branch may not have triggered any checks yet. Try pushing first.")
			return nil
		}
		fmt.Printf("\rWaiting for checks on '%s'...", info.Branch)
		time.Sleep(time.Second)
	}

	// Split checks into top-level runs and their sub-stages. A check named
	// "X (Stage Job)" is a sub-stage of a check named exactly "X" (Azure
	// Pipelines / GitHub Actions convention); anything else stays top-level.
	names := make(map[string]bool, len(checks))
	for _, c := range checks {
		names[c.Name] = true
	}

	var parents []Check
	children := make(map[string][]Check) // parent name → sub-stages, Name rewritten to the stage text
	for _, c := range checks {
		if parent, stage, ok := splitSubStage(c.Name, names); ok {
			c.Name = stage
			children[parent] = append(children[parent], c)
			continue
		}
		parents = append(parents, c)
	}

	var items []pickerItem
	for _, p := range parents {
		label := fmt.Sprintf("%-*s %s", checkNameWidth, p.Name, statusBadge(p.Status))
		if n := len(children[p.Name]); n > 0 {
			label += pickerDimStyle.Render(fmt.Sprintf(" +%d", n))
		}
		items = append(items, pickerItem{label: label, url: p.URL})
		for _, c := range children[p.Name] {
			label := pickerDimStyle.Render(fmt.Sprintf("├ %-*s", checkNameWidth-2, c.Name)) + " " + statusBadge(c.Status)
			items = append(items, pickerItem{label: label, url: c.URL, child: true})
		}
	}

	return runPicker(fmt.Sprintf("Checks — %s/%s (%s)", info.Owner, info.Repo, info.Branch), items)
}

func runPR(info RepoInfo, wait bool) error {
	var prs []PR
	for {
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
		prs = nil
		for _, pr := range prsFrom {
			seen[pr.Number] = true
			prs = append(prs, pr)
		}
		for _, pr := range prsInto {
			if !seen[pr.Number] {
				prs = append(prs, pr)
			}
		}

		if len(prs) > 0 {
			break
		}
		if !wait {
			fmt.Printf("No open pull requests found for branch '%s'.\n", info.Branch)
			return nil
		}
		fmt.Printf("\rWaiting for pull requests on '%s'...", info.Branch)
		time.Sleep(time.Second)
	}

	items := make([]pickerItem, len(prs))
	for i, pr := range prs {
		draft := ""
		if pr.Draft {
			draft = " [draft]"
		}
		label := fmt.Sprintf("#%-6d %s (%s → %s)%s", pr.Number, pr.Title, pr.HeadRef, pr.BaseRef, draft)
		items[i] = pickerItem{label: label, url: pr.URL}
	}

	return runPicker(fmt.Sprintf("Pull Requests — %s", info.Branch), items)
}

func main() {
	ciFlag := flag.Bool("ci", false, "Open CI check runs for the current branch")
	prFlag := flag.Bool("pr", false, "Open pull requests for the current branch")
	waitFlag := flag.Bool("wait", false, "Poll until check runs or pull requests are available (use with -ci or -pr)")

	flag.Usage = usage
	flag.Parse()

	if !*ciFlag && !*prFlag {
		usage()
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "Error: too many arguments; usage: gh-util -ci|-pr [branch]")
		os.Exit(1)
	}

	info, err := getRepoInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(args) == 1 {
		info.Branch = args[0]
	}

	if *ciFlag {
		err = runChecks(info, *waitFlag)
	} else {
		err = runPR(info, *waitFlag)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
