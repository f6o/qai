package cmd

import (
	"github.com/f6o/qai/i18n"
	"github.com/spf13/cobra"
)

var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: i18n.T("cmd.todo.short"),
}

func init() {
	rootCmd.AddCommand(todoCmd)
}
