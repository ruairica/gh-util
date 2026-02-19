# Pipeline CLI

Go CLI tool that opens Azure DevOps pipeline runs for the current GitHub repo/branch.

## Build & Run

```bash
go build -ldflags="-s -w" -o pipeline.exe .
go run .                    # quick test from this directory
go install -ldflags="-s -w" .  # install to GOPATH/bin
```

## Go Fix

After upgrading Go or making changes, run `go fix` to auto-modernize code:

```bash
# Preview changes without applying
go fix -diff ./...

# Apply fixes
go fix ./...
```

This replaces old patterns with modern equivalents (e.g. `interface{}` -> `any`, 3-clause for loops -> `range int`, `strings.Index` -> `strings.Cut`, if/else chains -> `min`/`max`). Run it after toolchain upgrades — start from a clean git state so changes are easy to review.

## Project Conventions

- Go 1.26+ — use modern language features (`any`, range-over-int, `min`/`max` builtins)
- External tools called via `os/exec`: `git`, `gh` (GitHub CLI)
- TUI selection via `charmbracelet/huh` (inline prompts, not full-screen)
- Status badges styled with `charmbracelet/lipgloss`
