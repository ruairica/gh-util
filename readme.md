A small Go CLI that quickly opens CI check runs or pull requests for your current Git branch.

Clone and run `go install -ldflags="-s -w" .` to add it to GOPATH. (requires gh cli)

Usage: gh-util [flags]

Flags:

  -ci   Open CI check runs for the current branch

  -pr   Open pull requests for the current branch
