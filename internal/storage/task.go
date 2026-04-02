package storage

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/f6o/qai/internal/model"
	"gopkg.in/yaml.v3"
)

type TaskStorage struct {
	filepath string
}

func NewTaskStorage(filepath string) *TaskStorage {
	return &TaskStorage{filepath: filepath}
}

func (s *TaskStorage) Load() ([]model.Task, error) {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Task{}, nil
		}
		return nil, err
	}

	var tasks []model.Task
	if err := yaml.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskStorage) Save(tasks []model.Task) error {
	data, err := yaml.Marshal(tasks)
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.filepath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(s.filepath, data, 0644)
}

func (s *TaskStorage) GetMaxID(tasks []model.Task) int {
	maxID := 0
	for _, t := range tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	return maxID
}

func (s *TaskStorage) Add(tasks []model.Task, task model.Task) ([]model.Task, error) {
	if task.ID == 0 {
		task.ID = s.GetMaxID(tasks) + 1
	}
	tasks = append(tasks, task)
	return tasks, s.Save(tasks)
}

func (s *TaskStorage) Update(tasks []model.Task, task model.Task) ([]model.Task, error) {
	idx := slices.IndexFunc(tasks, func(t model.Task) bool { return t.ID == task.ID })
	if idx == -1 {
		return tasks, nil
	}
	tasks[idx] = task
	return tasks, s.Save(tasks)
}

