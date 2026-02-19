# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# gh-util

Go CLI tool for GitHub repo utilities — quickly open Azure DevOps pipeline runs (`-p`) or pull requests (`-pr`) for the current branch.

## Commands

```bash
# Build
go build -ldflags="-s -w" -o gh-util.exe .

# Run
go run . -p                 # open pipelines for current branch
go run . -pr                # open PRs for current branch

# Install to GOPATH/bin (for testing from any repo)
go install -ldflags="-s -w" .

# Auto-modernize code after Go upgrades
go fix ./...
go fix -diff ./...          # preview changes
```

## Architecture

**Entry Point:** `main.go`
- Flag parsing (`-p` for pipelines, `-pr` for PRs)
- Routes to `runPipeline()` or `runPR()` handlers
- Single selection UI if multiple results found; auto-opens if single result

**Git Integration:** `git.go`
- `getRepoInfo()` extracts owner/repo from `git remote get-url origin`
- Parses both HTTPS (`https://github.com/owner/repo.git`) and SSH (`git@github.com:owner/repo.git`) URLs
- Falls back to `main` branch if current branch detection fails

**GitHub API:** `github.go`
- `fetchPipelines()` — queries GitHub API for Azure Pipelines check runs, deduplicates by name (keeps latest by `started_at`), filters to `azure-pipelines` app slug
- `fetchPRs()` — lists open PRs for current branch via `gh pr list` CLI

**Platform Support:** `open_*.go` (Windows/macOS/Linux)
- Opens URLs using platform-specific command (`start` / `open` / `xdg-open`)

**UI:** `main.go` status badges
- Green checkmark for `success`, red X for `failure`, orange arrows for `in_progress`
- Uses `charmbracelet/huh` for selection prompts (inline, not full-screen)

## Project Conventions

- Go 1.26+ — use modern language features (`any`, range-over-int, `min`/`max` builtins)
- External tools via `os/exec`: `git`, `gh` (GitHub CLI) — requires both installed and `gh` authenticated
- TUI via `charmbracelet/huh` (inline select prompts)
- Status styling via `charmbracelet/lipgloss` (colors and symbols)
- Strip binaries with `-ldflags="-s -w"` for size reduction
