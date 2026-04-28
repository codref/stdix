package main

import (
	"os"

	"github.com/codref/stdix/cmd/stdix/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
