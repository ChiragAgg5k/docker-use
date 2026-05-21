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
	_, _, _ = version, commit, date
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
