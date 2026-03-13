package storage

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/f6o/qai/internal/model"
	"gopkg.in/yaml.v3"
)

type TaskStorage struct {
	filepath     string
	doneFilepath string
}

func NewTaskStorage(filepath, doneFilepath string) *TaskStorage {
	return &TaskStorage{filepath: filepath, doneFilepath: doneFilepath}
}

func (s *TaskStorage) loadFile(path string) ([]model.Task, error) {
	data, err := os.ReadFile(path)
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

func (s *TaskStorage) writeFile(path string, tasks []model.Task) error {
	data, err := yaml.Marshal(tasks)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (s *TaskStorage) Load() ([]model.Task, error) {
	return s.loadFile(s.filepath)
}

func (s *TaskStorage) LoadDone() ([]model.Task, error) {
	return s.loadFile(s.doneFilepath)
}

func (s *TaskStorage) LoadAll() ([]model.Task, error) {
	active, err := s.loadFile(s.filepath)
	if err != nil {
		return nil, err
	}
	done, err := s.loadFile(s.doneFilepath)
	if err != nil {
		return nil, err
	}
	return append(active, done...), nil
}

func (s *TaskStorage) Save(tasks []model.Task) error {
	var active []model.Task
	var done []model.Task
	for _, t := range tasks {
		if t.Status == model.StatusDone {
			done = append(done, t)
		} else {
			active = append(active, t)
		}
	}

	// Write done.yaml first to minimize data loss on crash
	if len(done) > 0 {
		if err := s.appendDone(done); err != nil {
			return err
		}
	}

	return s.writeFile(s.filepath, active)
}

func (s *TaskStorage) appendDone(newDone []model.Task) error {
	existing, _ := s.loadFile(s.doneFilepath)
	idSet := make(map[int]bool)
	for _, t := range existing {
		idSet[t.ID] = true
	}
	for _, t := range newDone {
		if !idSet[t.ID] {
			existing = append(existing, t)
		}
	}
	return s.writeFile(s.doneFilepath, existing)
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
		allTasks, _ := s.LoadAll()
		task.ID = s.GetMaxID(allTasks) + 1
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

func (s *TaskStorage) FindByID(tasks []model.Task, id int) *model.Task {
	idx := slices.IndexFunc(tasks, func(t model.Task) bool { return t.ID == id })
	if idx == -1 {
		return nil
	}
	return &tasks[idx]
}

func (s *TaskStorage) FilterByStatus(tasks []model.Task, status model.Status) []model.Task {
	var result []model.Task
	for _, t := range tasks {
		if t.Status == status {
			result = append(result, t)
		}
	}
	return result
}

func (s *TaskStorage) FilterByParentID(tasks []model.Task, parentID int) []model.Task {
	var result []model.Task
	for _, t := range tasks {
		if t.ParentID != nil && *t.ParentID == parentID {
			result = append(result, t)
		}
	}
	return result
}

func (s *TaskStorage) FilterIdeas(tasks []model.Task) []model.Task {
	var result []model.Task
	for _, t := range tasks {
		if t.Status == model.StatusIdea {
			result = append(result, t)
		}
	}
	return result
}

func (s *TaskStorage) FilterTodos(tasks []model.Task) []model.Task {
	var result []model.Task
	for _, t := range tasks {
		if t.Status == model.StatusTodo || t.Status == model.StatusDoing {
			result = append(result, t)
		}
	}
	return result
}
