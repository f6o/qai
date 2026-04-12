package cmd

import (
	"github.com/f6o/qai/i18n"
	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: i18n.T("cmd.agent.short"),
}

func init() {
	rootCmd.AddCommand(agentCmd)
}
