package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"

	"github.com/f6o/qai/internal/model"
)

type LogStorage struct {
	filepath string
}

func NewLogStorage(filepath string) *LogStorage {
	return &LogStorage{filepath: filepath}
}

func (s *LogStorage) Load() ([]model.Log, error) {
	file, err := os.Open(s.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return []model.Log{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var logs []model.Log
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var log model.Log
		if err := json.Unmarshal(scanner.Bytes(), &log); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, scanner.Err()
}

func (s *LogStorage) Append(log model.Log) error {
	dir := filepath.Dir(s.filepath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(s.filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(log)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(data) + "\n")
	return err
}

func (s *LogStorage) GetMaxID(logs []model.Log) int {
	maxID := 0
	for _, l := range logs {
		if l.ID > maxID {
			maxID = l.ID
		}
	}
	return maxID
}

func (s *LogStorage) FilterByTodoID(logs []model.Log, todoID int) []model.Log {
	result := make([]model.Log, 0)
	for _, l := range logs {
		if l.TodoID == todoID {
			result = append(result, l)
		}
	}
	return result
}

func (s *LogStorage) FilterByDate(logs []model.Log, year int, month int, day int) []model.Log {
	result := make([]model.Log, 0)
	for _, l := range logs {
		if l.LoggedAt.Year() == year && int(l.LoggedAt.Month()) == month && l.LoggedAt.Day() == day {
			result = append(result, l)
		}
	}
	return result
}

func (s *LogStorage) FindByID(logs []model.Log, id int) *model.Log {
	idx := slices.IndexFunc(logs, func(l model.Log) bool { return l.ID == id })
	if idx == -1 {
		return nil
	}
	return &logs[idx]
}
