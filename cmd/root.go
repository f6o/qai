package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qai",
	Short: "qai - next generation task management tool",
}

func init() {
	rootCmd.SetOut(os.Stdout)
}

func Execute() error {
	return rootCmd.Execute()
}
