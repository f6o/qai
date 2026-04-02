package service

import (
	"context"
	"time"

	"github.com/f6o/qai/internal/model"
	"github.com/f6o/qai/internal/storage"
)

type LocalTaskService struct {
	store *storage.TaskStorage
}

func NewLocalTaskService(store *storage.TaskStorage) *LocalTaskService {
	return &LocalTaskService{store: store}
}

func (s *LocalTaskService) ListTasks(_ context.Context) ([]model.Task, error) {
	return s.store.Load()
}

func (s *LocalTaskService) AddTask(_ context.Context, task model.Task) (model.Task, error) {
	tasks, err := s.store.Load()
	if err != nil {
		return model.Task{}, err
	}
	tasks, err = s.store.Add(tasks, task)
	if err != nil {
		return model.Task{}, err
	}
	return tasks[len(tasks)-1], nil
}

func (s *LocalTaskService) UpdateTask(_ context.Context, task model.Task) (model.Task, error) {
	tasks, err := s.store.Load()
	if err != nil {
		return model.Task{}, err
	}
	tasks, err = s.store.Update(tasks, task)
	if err != nil {
		return model.Task{}, err
	}
	if t := model.FindTaskByID(tasks, task.ID); t != nil {
		return *t, nil
	}
	return task, nil
}

func (s *LocalTaskService) GetTask(_ context.Context, id int) (*model.Task, error) {
	tasks, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	return model.FindTaskByID(tasks, id), nil
}

type LocalLogService struct {
	store *storage.LogStorage
}

func NewLocalLogService(store *storage.LogStorage) *LocalLogService {
	return &LocalLogService{store: store}
}

func (s *LocalLogService) AppendLog(_ context.Context, log model.Log) (model.Log, error) {
	logs, err := s.store.Load()
	if err != nil {
		return model.Log{}, err
	}
	log.ID = s.store.GetMaxID(logs) + 1
	log.LoggedAt = time.Now()
	if err := s.store.Append(log); err != nil {
		return model.Log{}, err
	}
	return log, nil
}

func (s *LocalLogService) ListLogs(_ context.Context, opts LogListOptions) ([]model.Log, error) {
	logs, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	if opts.EventType != nil {
		logs = model.FilterLogsByEventType(logs, *opts.EventType)
	}
	if opts.TodoID != nil {
		logs = model.FilterLogsByTodoID(logs, *opts.TodoID)
	}
	if opts.Year != nil && opts.Month != nil && opts.Day != nil {
		logs = model.FilterLogsByDate(logs, *opts.Year, *opts.Month, *opts.Day)
	}
	return logs, nil
}
