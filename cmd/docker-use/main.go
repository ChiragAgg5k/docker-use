package main

import (
	"os"

	"github.com/chiragagg5k/docker-use/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
