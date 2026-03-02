package cmd

import (
	"github.com/spf13/cobra"
)

var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: "Manage todos",
}

func init() {
	rootCmd.AddCommand(todoCmd)
}
