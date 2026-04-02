package cmd

import (
	"fmt"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/config"
	"github.com/f6o/qai/internal/service"
	"github.com/f6o/qai/internal/storage"
)

type AppContext struct {
	Config *config.Config
	Tasks  service.TaskService
	Logs   service.LogService
}

func NewAppContext() (*AppContext, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("error.config_load"), err)
	}

	if err := cfg.EnsureDirectories(); err != nil {
		return nil, fmt.Errorf(i18n.T("error.create_dirs"), err)
	}

	var tasks service.TaskService
	var logs service.LogService

	switch cfg.Server.Mode {
	case "server":
		conn, err := service.DialUnix(cfg.Server.SocketPath)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to qai server: %w", err)
		}
		tasks, logs = service.NewRemoteServices(conn)
	default:
		ts := storage.NewTaskStorage(cfg.Data.Todofile)
		ls := storage.NewLogStorage(cfg.Data.Logfile)
		tasks = service.NewLocalTaskService(ts)
		logs = service.NewLocalLogService(ls)
	}

	return &AppContext{
		Config: cfg,
		Tasks:  tasks,
		Logs:   logs,
	}, nil
}
