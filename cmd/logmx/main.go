package main

import (
	"fmt"
	"os"

	"github.com/lucasnevespereira/logmx/cmd/logmx/commands"
)

func main() {
	if err := commands.Root().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
