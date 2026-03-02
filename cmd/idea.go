package cmd

import (
	"github.com/spf13/cobra"
)

var ideaCmd = &cobra.Command{
	Use:   "idea",
	Short: "Manage ideas",
}

func init() {
	rootCmd.AddCommand(ideaCmd)
}
