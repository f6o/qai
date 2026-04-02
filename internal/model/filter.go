package model

import "slices"

func FilterTodos(tasks []Task) []Task {
	var result []Task
	for _, t := range tasks {
		if t.Status == StatusTodo || t.Status == StatusDoing {
			result = append(result, t)
		}
	}
	return result
}

func FilterIdeas(tasks []Task) []Task {
	var result []Task
	for _, t := range tasks {
		if t.Status == StatusIdea {
			result = append(result, t)
		}
	}
	return result
}

func FilterByParentID(tasks []Task, parentID int) []Task {
	var result []Task
	for _, t := range tasks {
		if t.ParentID != nil && *t.ParentID == parentID {
			result = append(result, t)
		}
	}
	return result
}

func FilterByStatus(tasks []Task, status Status) []Task {
	var result []Task
	for _, t := range tasks {
		if t.Status == status {
			result = append(result, t)
		}
	}
	return result
}

func FindTaskByID(tasks []Task, id int) *Task {
	idx := slices.IndexFunc(tasks, func(t Task) bool { return t.ID == id })
	if idx == -1 {
		return nil
	}
	return &tasks[idx]
}

func FilterLogsByEventType(logs []Log, eventType EventType) []Log {
	result := make([]Log, 0)
	for _, l := range logs {
		if l.EffectiveEventType() == eventType {
			result = append(result, l)
		}
	}
	return result
}

func FilterLogsByDate(logs []Log, year int, month int, day int) []Log {
	result := make([]Log, 0)
	for _, l := range logs {
		if l.LoggedAt.Year() == year && int(l.LoggedAt.Month()) == month && l.LoggedAt.Day() == day {
			result = append(result, l)
		}
	}
	return result
}

func FilterLogsByTodoID(logs []Log, todoID int) []Log {
	result := make([]Log, 0)
	for _, l := range logs {
		if l.TodoID == todoID {
			result = append(result, l)
		}
	}
	return result
}

func FindLogByID(logs []Log, id int) *Log {
	idx := slices.IndexFunc(logs, func(l Log) bool { return l.ID == id })
	if idx == -1 {
		return nil
	}
	return &logs[idx]
}
