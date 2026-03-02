package cmd

import (
	"fmt"

	"github.com/f6o/qai/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize qai configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Default()

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		cmd.Println("Initialized qai at:", cfg.Data.Todofile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
