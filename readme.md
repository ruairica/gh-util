A small Go CLI that quickly opens CI check runs or pull requests for your current Git branch.

Clone and run `go install -ldflags="-s -w" .` to add it to GOPATH. (requires gh cli)

Usage: gh-util [flags] [branch]

Flags:

  -ci [branch]   Open CI check runs (current branch by default; pass a branch name to inspect another without checking it out)

  -wait          Poll until CI check runs appear for the branch, then open them (use with -ci when checks haven't started yet; combines with a branch name, e.g. `gh-util -ci -wait some-branch`)

  -pr [branch]   Open pull requests (current branch by default; pass a branch name e.g. `main` to see the review queue targeting it)
