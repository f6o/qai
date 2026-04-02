package service

import (
	"context"

	"github.com/f6o/qai/internal/model"
)

type TaskService interface {
	ListTasks(ctx context.Context) ([]model.Task, error)
	AddTask(ctx context.Context, task model.Task) (model.Task, error)
	UpdateTask(ctx context.Context, task model.Task) (model.Task, error)
	GetTask(ctx context.Context, id int) (*model.Task, error)
}

type LogService interface {
	AppendLog(ctx context.Context, log model.Log) (model.Log, error)
	ListLogs(ctx context.Context, opts LogListOptions) ([]model.Log, error)
}

type LogListOptions struct {
	EventType *model.EventType
	TodoID    *int
	Year      *int
	Month     *int
	Day       *int
}
