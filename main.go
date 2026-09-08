package main

import (
	"os"

	"github.com/f6o/qai/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
