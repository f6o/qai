package cmd

import (
	"fmt"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: i18n.T("cmd.init.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Default()

		if err := cfg.Save(); err != nil {
			return fmt.Errorf(i18n.T("cmd.init.error"), err)
		}

		cmd.Println(i18n.T("cmd.init.success", cfg.Data.Todofile))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
