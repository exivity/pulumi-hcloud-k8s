//go:build tools
// +build tools

package main

import (
	// https://go.dev/blog/vuln
	_ "golang.org/x/vuln/cmd/govulncheck"
	// hcloud CLI
	_ "github.com/hetznercloud/cli/cmd/hcloud"
)
