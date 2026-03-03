package cmd

import (
	"os"

	"github.com/f6o/qai/i18n"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "qai",
	Short: i18n.T("cmd.root.short"),
}

func init() {
	rootCmd.SetOut(os.Stdout)
}

func Execute() error {
	return rootCmd.Execute()
}
