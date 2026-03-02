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

type Log struct {
	ID       int       `json:"id"`
	TodoID   int       `json:"todo_id"`
	Content  string    `json:"content"`
	Duration int       `json:"duration"`
	LoggedAt time.Time `json:"logged_at"`
}
