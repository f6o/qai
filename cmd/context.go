package cmd

import (
	"fmt"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/config"
	"github.com/f6o/qai/internal/storage"
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
		TaskStore: storage.NewTaskStorage(cfg.Data.Todofile, cfg.Data.Donefile),
		LogStore:  storage.NewLogStorage(cfg.Data.Logfile),
	}, nil
}
