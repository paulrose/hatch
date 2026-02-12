package main

import (
	"os"

	"github.com/paulrose/hatch/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
