package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/flock"
	"github.com/f6o/qai/internal/pomo"
	"github.com/spf13/cobra"
)

var timerCmd = &cobra.Command{
	Use:     "timer",
	Aliases: []string{"pomo"},
	Short:   i18n.T("cmd.timer.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		lockPath := filepath.Join(filepath.Dir(ctx.Config.Data.Todofile), "timer.lock")
		fl := flock.New(lockPath)
		locked, err := fl.TryLock()
		if err != nil {
			return fmt.Errorf(i18n.T("cmd.timer.error_lock"), err)
		}
		if !locked {
			return errors.New(i18n.T("cmd.timer.error_locked"))
		}
		defer fl.Unlock()

		m := pomo.NewModel(ctx.Config, ctx.Tasks, ctx.Logs)
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
