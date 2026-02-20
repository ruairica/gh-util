A small Go CLI that quickly opens CI pipeline runs or pull requests for your current Git branch.

Clone and run `go install -ldflags="-s -w" .` to add it to GOPATH. (requires gh cli)

Usage: gh-util [flags]

Flags:

  -p, --pipeline   Open pipeline runs for the current branch

  -pr              Open pull requests for the current branch
