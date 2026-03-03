package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/config"
	"github.com/f6o/qai/internal/pomo"
	"github.com/f6o/qai/internal/storage"
	"github.com/spf13/cobra"
)

var pomoCmd = &cobra.Command{
	Use:   "pomo",
	Short: i18n.T("cmd.pomo.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf(i18n.T("cmd.pomo.error_config"), err)
		}

		ts := storage.NewTaskStorage(cfg.Data.Todofile)
		ls := storage.NewLogStorage(cfg.Data.Logfile)

		m := pomo.NewModel(cfg, ts, ls)
		p := tea.NewProgram(&m, tea.WithInput(cmd.InOrStdin()), tea.WithOutput(cmd.OutOrStdout()))

		if _, err := p.Run(); err != nil {
			return fmt.Errorf(i18n.T("cmd.pomo.error_run"), err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pomoCmd)
}
