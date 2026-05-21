package main

import (
	"os"

	"github.com/chiragagg5k/docker-use/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := cli.NewRootCommand(version).Execute(); err != nil {
		os.Exit(1)
	}
}
