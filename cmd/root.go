package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qai",
	Short: "qai - next generation task management tool",
}

func Execute() error {
	return rootCmd.Execute()
}
