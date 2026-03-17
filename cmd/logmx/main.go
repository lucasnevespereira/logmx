package main

import (
	"fmt"
	"os"

	"github.com/lucasnevespereira/logmx/cmd/logmx/commands"
)

var version = "dev"

func main() {
	if err := commands.Root(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
