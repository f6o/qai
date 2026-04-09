package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/config"
	"github.com/f6o/qai/internal/flock"
	"github.com/f6o/qai/internal/pomo"
	"github.com/f6o/qai/internal/storage"
	"github.com/spf13/cobra"
)

type AppContext struct {
	Config    *config.Config
	TaskStore *storage.TaskStorage
	LogStore  *storage.LogStorage
}

func NewAppContext() (*AppContext, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("error.config_load"), err)
	}

	if err := cfg.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf(i18n.T("error.create_dirs"), err)
	}

	return &AppContext{
		Config:    cfg,
		TaskStore: storage.NewTaskStorage(cfg.Data.Todofile),
		LogStore:  storage.NewLogStorage(cfg.Data.Logfile),
	}, nil
}

func (ctx *AppContext) RunPomodoro(cmd *cobra.Command, taskID int) error {
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

	m := pomo.NewModel(ctx.Config, ctx.TaskStore, ctx.LogStore)
	if taskID > 0 {
		m.AutoStartTaskID = taskID
	}
	p := tea.NewProgram(&m, tea.WithInput(cmd.InOrStdin()), tea.WithOutput(cmd.OutOrStdout()))

	if _, err := p.Run(); err != nil {
		return fmt.Errorf(i18n.T("cmd.timer.error_run"), err)
	}
	return nil
}
