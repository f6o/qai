package pomo

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/config"
	"github.com/f6o/qai/internal/markdown"
	"github.com/f6o/qai/internal/model"
	"github.com/f6o/qai/internal/storage"
)

type State int

const (
	StateSelectTask State = iota
	StateFocus
	StateBreakChoice
	StateBreak
	StateBreakDone
)

type Model struct {
	Tasks       []model.Task
	TaskStore   *storage.TaskStorage
	LogStore    *storage.LogStorage
	Config      *config.Config
	MarkdownGen *markdown.Generator

	CurrentState State
	SelectedIdx  int
	FocusedTask  *model.Task

	StartTime   time.Time
	CompletedAt time.Time
	TimeLeft    time.Duration
	IsPaused    bool
	PausedAt    time.Time

	SessionType       string
	CompletedSessions int
}

func NewModel(cfg *config.Config, ts *storage.TaskStorage, ls *storage.LogStorage) Model {
	return Model{
		Tasks:       nil,
		TaskStore:   ts,
		LogStore:    ls,
		Config:      cfg,
		MarkdownGen: markdown.NewGenerator(cfg.Data.MarkdownDir),

		CurrentState: StateSelectTask,
		SelectedIdx:  0,
		FocusedTask:  nil,

		StartTime:   time.Time{},
		CompletedAt: time.Time{},
		TimeLeft:    0,
		IsPaused:    false,
		PausedAt:    time.Time{},

		SessionType:       "work",
		CompletedSessions: 0,
	}
}

func (m *Model) Init() tea.Cmd {
	loadTasks := func() tea.Msg {
		tasks, err := m.TaskStore.Load()
		if err != nil {
			return tea.Msg(fmt.Sprintf("Error: %v", err))
		}
		return tasks
	}
	return loadTasks
}

type TickMsg time.Time

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []model.Task:
		m.Tasks = msg
		if m.FocusedTask != nil {
			for i := range m.Tasks {
				if m.Tasks[i].ID == m.FocusedTask.ID {
					m.FocusedTask = &m.Tasks[i]
					break
				}
			}
		}
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	case TickMsg:
		return m.handleTick(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.CurrentState {
	case StateSelectTask:
		return m.handleSelectTask(msg)
	case StateFocus:
		return m.handleFocus(msg)
	case StateBreakChoice:
		return m.handleBreakChoice(msg)
	case StateBreak:
		return m.handleBreak(msg)
	case StateBreakDone:
		return m.handleBreakDone(msg)
	}
	return m, nil
}

func (m *Model) getSortedTodos() []model.Task {
	todos := m.TaskStore.FilterTodos(m.Tasks)
	sort.Slice(todos, func(i, j int) bool {
		if todos[i].Priority != todos[j].Priority {
			return todos[i].Priority > todos[j].Priority
		}
		return todos[i].ID < todos[j].ID
	})
	return todos
}

func (m *Model) handleSelectTask(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	todos := m.getSortedTodos()

	switch msg.String() {
	case "up", "k":
		if m.SelectedIdx > 0 {
			m.SelectedIdx--
		}
	case "down", "j":
		if m.SelectedIdx < len(todos)-1 {
			m.SelectedIdx++
		}
	case "+", "-", "<", ">":
		if len(todos) > 0 {
			task := todos[m.SelectedIdx]
			switch msg.String() {
			case "+":
				task.Priority++
			case "-":
				task.Priority--
			case ">":
				task.Priority += 5
			case "<":
				task.Priority -= 5
			}
			m.Tasks, _ = m.TaskStore.Update(m.Tasks, task)

			// Re-sort and find new index to keep cursor on the same task
			newTodos := m.getSortedTodos()
			for i, t := range newTodos {
				if t.ID == task.ID {
					m.SelectedIdx = i
					break
				}
			}
		}
	case "enter":
		if len(todos) > 0 {
			task := todos[m.SelectedIdx]
			task.Status = model.StatusDoing
			task.StartedAt = time.Now()
			m.Tasks, _ = m.TaskStore.Update(m.Tasks, task)

			// Find the task in updated m.Tasks to set FocusedTask
			for i := range m.Tasks {
				if m.Tasks[i].ID == task.ID {
					m.FocusedTask = &m.Tasks[i]
					break
				}
			}

			m.CurrentState = StateFocus
			m.SessionType = "work"
			m.StartTime = time.Now()
			m.CompletedAt = time.Time{}
			m.TimeLeft = time.Duration(m.Config.Pomodoro.WorkMinutes) * time.Minute
			m.IsPaused = false
			m.saveMarkdown()
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return TickMsg(t) })
		}
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) handleFocus(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "p":
		if !m.IsPaused {
			m.IsPaused = true
			m.PausedAt = time.Now()
		}
	case "r":
		if m.IsPaused {
			m.IsPaused = false
			pausedDuration := time.Since(m.PausedAt)
			m.StartTime = m.StartTime.Add(pausedDuration)
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return TickMsg(t) })
		}
	case "s":
		m.CurrentState = StateBreakChoice
	case "q", "ctrl+c", "esc":
		m.CompletedAt = time.Now()
		m.CurrentState = StateBreakChoice
	}
	return m, nil
}

func (m *Model) handleBreakChoice(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "b":
		m.CurrentState = StateBreak
		m.SessionType = "break"
		m.StartTime = time.Now()
		m.CompletedAt = time.Time{}
		m.TimeLeft = time.Duration(m.Config.Pomodoro.BreakMinutes) * time.Minute
		m.IsPaused = false
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return TickMsg(t) })
	case "s":
		m.CurrentState = StateBreakDone
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) handleBreak(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s":
		m.CurrentState = StateBreakDone
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) handleBreakDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		m.CurrentState = StateFocus
		m.SessionType = "work"
		m.StartTime = time.Now()
		m.CompletedAt = time.Time{}
		m.TimeLeft = time.Duration(m.Config.Pomodoro.WorkMinutes) * time.Minute
		m.IsPaused = false
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return TickMsg(t) })
	case "n":
		m.CurrentState = StateSelectTask
		m.SelectedIdx = 0
		m.FocusedTask = nil
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) handleTick(_ TickMsg) (tea.Model, tea.Cmd) {
	if m.IsPaused {
		return m, nil
	}

	elapsed := time.Since(m.StartTime)
	m.TimeLeft = time.Duration(m.Config.Pomodoro.WorkMinutes)*time.Minute - elapsed
	if m.SessionType == "break" {
		m.TimeLeft = time.Duration(m.Config.Pomodoro.BreakMinutes)*time.Minute - elapsed
	}

	if m.TimeLeft <= 0 {
		switch m.CurrentState {
		case StateFocus:
			m.CompletedSessions++
			m.CompletedAt = time.Now()
			m.saveLog(m.FocusedTask.ID, m.FocusedTask.Title, m.Config.Pomodoro.WorkMinutes)
			m.CurrentState = StateBreakChoice
		case StateBreak:
			m.CurrentState = StateBreakDone
		}
		return m, nil
	}

	return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return TickMsg(t) })
}

func (m *Model) saveLog(todoID int, content string, duration int) {
	logs, _ := m.LogStore.Load()
	log := model.Log{
		ID:       m.LogStore.GetMaxID(logs) + 1,
		TodoID:   todoID,
		Content:  content,
		Duration: duration,
		LoggedAt: time.Now(),
	}
	m.LogStore.Append(log)
}

func (m *Model) saveMarkdown() {
	m.saveMarkdownAt(time.Now())
}

func (m *Model) saveMarkdownAt(t time.Time) {
	m.MarkdownGen.Save(m.Tasks, t)
}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	subtleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

func (m *Model) View() string {
	switch m.CurrentState {
	case StateSelectTask:
		return m.viewSelectTask()
	case StateFocus:
		return m.viewFocus()
	case StateBreakChoice:
		return m.viewBreakChoice()
	case StateBreak:
		return m.viewBreak()
	case StateBreakDone:
		return m.viewBreakDone()
	}
	return ""
}

func (m *Model) viewSelectTask() string {
	var s string
	s += titleStyle.Render(i18n.T("pomo.select_task")) + "\n\n"

	todos := m.getSortedTodos()
	for i, t := range todos {
		prefix := "  "
		if i == m.SelectedIdx {
			prefix = "> "
			s += selectedStyle.Render(fmt.Sprintf("%s[%d] (P%d) [%s] %s", prefix, t.ID, t.Priority, t.Status, t.Title)) + "\n"
		} else {
			s += fmt.Sprintf("%s[%d] (P%d) [%s] %s\n", prefix, t.ID, t.Priority, t.Status, t.Title)
		}
	}

	s += "\n" + subtleStyle.Render(i18n.T("pomo.select_task_hint"))
	return s
}

func (m *Model) viewFocus() string {
	var s string
	s += titleStyle.Render(i18n.T("pomo.focusing_on", m.FocusedTask.ID, m.FocusedTask.Title)) + "\n\n"

	s += subtleStyle.Render(fmt.Sprintf("Started: %s", m.StartTime.Format("15:04"))) + "\n\n"

	elapsed := time.Since(m.StartTime)
	totalDuration := time.Duration(m.Config.Pomodoro.WorkMinutes) * time.Minute
	if elapsed > totalDuration {
		elapsed = totalDuration
	}
	progress := float64(elapsed) / float64(totalDuration)
	barWidth := 20
	filled := int(progress * float64(barWidth))
	var bar strings.Builder
	for i := range barWidth {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	s += checkboxStyle.Render(fmt.Sprintf("[%s] %s / %s", bar.String(), formatDuration(elapsed), formatDuration(totalDuration))) + "\n\n"

	if m.IsPaused {
		s += subtleStyle.Render(i18n.T("pomo.focus_paused"))
	} else {
		s += subtleStyle.Render(i18n.T("pomo.focus_running"))
	}
	return s
}

func (m *Model) viewBreakChoice() string {
	var s string
	s += titleStyle.Render(i18n.T("pomo.break_choice_title")) + "\n\n"

	if m.FocusedTask != nil {
		s += subtleStyle.Render(i18n.T("pomo.current_task", m.FocusedTask.ID, m.FocusedTask.Title)) + "\n\n"
	}

	s += subtleStyle.Render(fmt.Sprintf("Completed at: %s", m.CompletedAt.Format("15:04"))) + "\n\n"
	s += i18n.T("pomo.break_choice_options")
	return s
}

func (m *Model) viewBreak() string {
	var s string
	s += titleStyle.Render(i18n.T("pomo.break_title")) + "\n\n"

	if m.FocusedTask != nil {
		s += subtleStyle.Render(i18n.T("pomo.current_task", m.FocusedTask.ID, m.FocusedTask.Title)) + "\n\n"
	}

	elapsed := time.Since(m.StartTime)
	totalDuration := time.Duration(m.Config.Pomodoro.BreakMinutes) * time.Minute
	if elapsed > totalDuration {
		elapsed = totalDuration
	}
	progress := float64(elapsed) / float64(totalDuration)
	barWidth := 20
	filled := int(progress * float64(barWidth))
	var bar strings.Builder
	for i := range barWidth {
		if i < filled {
			bar.WriteString("█")
		} else {
			bar.WriteString("░")
		}
	}

	s += checkboxStyle.Render(fmt.Sprintf("[%s] %s / %s", bar.String(), formatDuration(elapsed), formatDuration(totalDuration))) + "\n\n"

	s += subtleStyle.Render(i18n.T("pomo.break_skip"))
	return s
}

func (m *Model) viewBreakDone() string {
	var s string
	s += titleStyle.Render(i18n.T("pomo.break_done_title")) + "\n\n"

	if m.FocusedTask != nil {
		s += subtleStyle.Render(i18n.T("pomo.current_task", m.FocusedTask.ID, m.FocusedTask.Title)) + "\n\n"
	}

	s += i18n.T("pomo.break_done_options")
	return s
}

func formatDuration(d time.Duration) string {
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}
