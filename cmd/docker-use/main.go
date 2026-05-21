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
	cli.Version = version
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
