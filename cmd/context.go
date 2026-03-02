package cmd

import (
	"fmt"

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
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf("failed to create directories: %w", err)
	}

	return &AppContext{
		Config:    cfg,
		TaskStore: storage.NewTaskStorage(cfg.Data.Todofile),
		LogStore:  storage.NewLogStorage(cfg.Data.Logfile),
	}, nil
}
