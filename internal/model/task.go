package model

import (
	"time"
)

type Status string

const (
	StatusIdea  Status = "idea"
	StatusTodo  Status = "todo"
	StatusDoing Status = "doing"
	StatusDone  Status = "done"
)

type Task struct {
	ID        int       `yaml:"id" json:"id"`
	Title     string    `yaml:"title" json:"title"`
	Status    Status    `yaml:"status" json:"status"`
	Priority  int       `yaml:"priority" json:"priority"`
	ParentID  *int      `yaml:"parent_id,omitempty" json:"parent_id,omitempty"`
	StartedAt time.Time `yaml:"started_at,omitempty" json:"started_at,omitempty"`
	CreatedAt time.Time `yaml:"created_at" json:"created_at"`
}

type EventType string

const (
	EventFocusComplete EventType = "focus_complete"
	EventFocusSkip     EventType = "focus_skip"
	EventFocusQuit     EventType = "focus_quit"
	EventTaskCreate    EventType = "task_create"
	EventStatusChange  EventType = "status_change"
)

type Log struct {
	ID         int       `json:"id"`
	TodoID     int       `json:"todo_id"`
	Content    string    `json:"content,omitempty"`
	Duration   *int      `json:"duration,omitempty"`
	LoggedAt   time.Time `json:"logged_at"`
	EventType  EventType `json:"event_type,omitempty"`
	FromStatus Status    `json:"from_status,omitempty"`
	ToStatus   Status    `json:"to_status,omitempty"`
}

func (l Log) EffectiveEventType() EventType {
	if l.EventType == "" {
		return EventFocusComplete
	}
	return l.EventType
}

func IntPtr(v int) *int {
	return &v
}
