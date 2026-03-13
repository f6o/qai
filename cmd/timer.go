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

var timerCmd = &cobra.Command{
	Use:     "timer",
	Aliases: []string{"pomo"},
	Short:   i18n.T("cmd.timer.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf(i18n.T("cmd.timer.error_config"), err)
		}

		ts := storage.NewTaskStorage(cfg.Data.Todofile, cfg.Data.Donefile)
		ls := storage.NewLogStorage(cfg.Data.Logfile)

		m := pomo.NewModel(cfg, ts, ls)
		p := tea.NewProgram(&m, tea.WithInput(cmd.InOrStdin()), tea.WithOutput(cmd.OutOrStdout()))

		if _, err := p.Run(); err != nil {
			return fmt.Errorf(i18n.T("cmd.timer.error_run"), err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(timerCmd)
}
