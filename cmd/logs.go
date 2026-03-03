package cmd

import (
	"github.com/f6o/qai/i18n"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: i18n.T("cmd.logs.short"),
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
