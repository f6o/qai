package cmd

import (
	"github.com/f6o/qai/i18n"
	"github.com/spf13/cobra"
)

var ideaCmd = &cobra.Command{
	Use:   "idea",
	Short: i18n.T("cmd.idea.short"),
}

func init() {
	rootCmd.AddCommand(ideaCmd)
}
