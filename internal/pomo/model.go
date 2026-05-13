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
	"github.com/f6o/qai/internal/pomo/alert"
	"github.com/f6o/qai/internal/storage"
)

type State int

const (
	StateSelectTask State = iota
	StateFocus
	StateBreakChoice
	StateBreak
	StateBreakDone
	StateNewTaskInput
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

	AutoStartTaskID int

	SearchMode  bool
	SearchQuery string

	NewTaskTitle string
}

const (
	actionReminderDelay = 3 * time.Minute
	actionReminderID    = "pomo-action-reminder"
)

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

func (m *Model) cancelActionReminder() tea.Cmd {
	return alert.Cancel(actionReminderID)
}

func (m *Model) enterActionPromptState(state State) tea.Cmd {
	m.CurrentState = state
	if !m.Config.Pomodoro.Notify {
		return nil
	}
	return alert.Schedule(actionReminderID, "qai", m.actionReminderText(state), actionReminderDelay, actionReminderDelay)
}

func (m *Model) actionReminderText(state State) string {
	switch state {
	case StateBreakChoice:
		return i18n.T("pomo.notify_break_choice_reminder")
	case StateBreakDone:
		return i18n.T("pomo.notify_break_done_reminder")
	default:
		return i18n.T("pomo.notify_work_complete")
	}
}

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
		if m.AutoStartTaskID > 0 && m.CurrentState == StateSelectTask {
			return m.autoStartTask()
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
	case StateNewTaskInput:
		return m.handleNewTaskInput(msg)
	}
	return m, nil
}

func (m *Model) startTask(taskID int) (tea.Model, tea.Cmd) {
	for i := range m.Tasks {
		if m.Tasks[i].ID == taskID {
			task := m.Tasks[i]
			fromStatus := task.Status
			task.Status = model.StatusDoing
			task.StartedAt = time.Now()
			m.Tasks, _ = m.TaskStore.Update(m.Tasks, task)

			m.LogStore.AppendNew(model.Log{
				TodoID:     task.ID,
				EventType:  model.EventStatusChange,
				FromStatus: fromStatus,
				ToStatus:   model.StatusDoing,
			})

			for j := range m.Tasks {
				if m.Tasks[j].ID == taskID {
					m.FocusedTask = &m.Tasks[j]
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
			return m, tea.Batch(
				m.cancelActionReminder(),
				tea.Tick(time.Second, func(t time.Time) tea.Msg { return TickMsg(t) }),
			)
		}
	}

	return m, nil
}

func (m *Model) autoStartTask() (tea.Model, tea.Cmd) {
	taskID := m.AutoStartTaskID
	m.AutoStartTaskID = 0
	return m.startTask(taskID)
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

func (m *Model) getVisibleTodos() []model.Task {
	todos := m.getSortedTodos()
	if m.SearchQuery == "" {
		return todos
	}
	q := strings.ToLower(m.SearchQuery)
	filtered := todos[:0:0]
	for _, t := range todos {
		if strings.Contains(strings.ToLower(t.Title), q) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func (m *Model) clampSelectedIdx(n int) {
	if n == 0 {
		m.SelectedIdx = 0
		return
	}
	if m.SelectedIdx >= n {
		m.SelectedIdx = n - 1
	}
	if m.SelectedIdx < 0 {
		m.SelectedIdx = 0
	}
}

func (m *Model) handleSelectTask(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.SearchMode {
		return m.handleSearchInput(msg)
	}

	todos := m.getVisibleTodos()

	switch msg.String() {
	case "/":
		m.SearchMode = true
		return m, nil
	case "ctrl+p", "up", "k":
		if m.SelectedIdx > 0 {
			m.SelectedIdx--
		}
	case "ctrl+n", "down", "j":
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
			newTodos := m.getVisibleTodos()
			for i, t := range newTodos {
				if t.ID == task.ID {
					m.SelectedIdx = i
					break
				}
			}
			m.clampSelectedIdx(len(newTodos))
		}
	case "enter":
		if len(todos) > 0 {
			task := todos[m.SelectedIdx]
			return m.startTask(task.ID)
		}
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) completeFocusedTask() {
	if m.FocusedTask == nil {
		return
	}
	fromStatus := m.FocusedTask.Status
	m.FocusedTask.Status = model.StatusDone
	m.Tasks, _ = m.TaskStore.Update(m.Tasks, *m.FocusedTask)
	m.LogStore.AppendNew(model.Log{
		TodoID:     m.FocusedTask.ID,
		EventType:  model.EventStatusChange,
		FromStatus: fromStatus,
		ToStatus:   model.StatusDone,
	})
	m.saveMarkdown()
}

func (m *Model) enterNewTaskInput() tea.Cmd {
	m.completeFocusedTask()
	m.CurrentState = StateNewTaskInput
	m.NewTaskTitle = ""
	m.FocusedTask = nil
	m.CompletedAt = time.Time{}
	return m.cancelActionReminder()
}

func (m *Model) addAndStartNewTask() (tea.Model, tea.Cmd) {
	title := strings.TrimSpace(m.NewTaskTitle)
	if title == "" {
		return m, nil
	}

	task := model.Task{
		Title:     title,
		Status:    model.StatusTodo,
		Priority:  m.Config.Task.DefaultPriority,
		ParentID:  nil,
		CreatedAt: time.Now(),
	}

	var err error
	m.Tasks, err = m.TaskStore.Add(m.Tasks, task)
	if err != nil {
		return m, nil
	}

	task = m.Tasks[len(m.Tasks)-1]
	m.LogStore.AppendNew(model.Log{
		TodoID:    task.ID,
		Content:   task.Title,
		EventType: model.EventTaskCreate,
	})
	m.NewTaskTitle = ""
	return m.startTask(task.ID)
}

func (m *Model) handleNewTaskInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		return m.addAndStartNewTask()
	case tea.KeyEsc, tea.KeyCtrlC:
		m.CurrentState = StateSelectTask
		m.NewTaskTitle = ""
		m.SelectedIdx = 0
		return m, nil
	case tea.KeyBackspace:
		if len(m.NewTaskTitle) > 0 {
			title := []rune(m.NewTaskTitle)
			m.NewTaskTitle = string(title[:len(title)-1])
		}
		return m, nil
	case tea.KeyRunes:
		m.NewTaskTitle += string(msg.Runes)
		return m, nil
	case tea.KeySpace:
		m.NewTaskTitle += " "
		return m, nil
	}
	return m, nil
}

func (m *Model) handleSearchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.SearchMode = false
		m.clampSelectedIdx(len(m.getVisibleTodos()))
		return m, nil
	case tea.KeyEsc, tea.KeyCtrlC:
		m.SearchMode = false
		m.SearchQuery = ""
		m.SelectedIdx = 0
		return m, nil
	case tea.KeyBackspace:
		if len(m.SearchQuery) > 0 {
			q := []rune(m.SearchQuery)
			m.SearchQuery = string(q[:len(q)-1])
		}
		m.SelectedIdx = 0
		return m, nil
	case tea.KeyRunes:
		m.SearchQuery += string(msg.Runes)
		m.SelectedIdx = 0
		return m, nil
	case tea.KeySpace:
		m.SearchQuery += " "
		m.SelectedIdx = 0
		return m, nil
	}
	return m, nil
}

func (m *Model) matchesSearch(title string) bool {
	if m.SearchQuery == "" {
		return false
	}
	return strings.Contains(strings.ToLower(title), strings.ToLower(m.SearchQuery))
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
		if m.FocusedTask != nil {
			elapsed := int(time.Since(m.StartTime).Minutes())
			m.LogStore.AppendNew(model.Log{
				TodoID:    m.FocusedTask.ID,
				Duration:  model.IntPtr(elapsed),
				EventType: model.EventFocusSkip,
			})
		}
		return m, m.enterActionPromptState(StateBreakChoice)
	case "q", "ctrl+c", "esc":
		if m.FocusedTask != nil {
			elapsed := int(time.Since(m.StartTime).Minutes())
			m.LogStore.AppendNew(model.Log{
				TodoID:    m.FocusedTask.ID,
				Duration:  model.IntPtr(elapsed),
				EventType: model.EventFocusQuit,
			})
		}
		m.CompletedAt = time.Now()
		return m, m.enterActionPromptState(StateBreakChoice)
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
		return m, tea.Batch(
			m.cancelActionReminder(),
			tea.Tick(time.Second, func(t time.Time) tea.Msg { return TickMsg(t) }),
		)
	case "d":
		m.completeFocusedTask()
		m.CurrentState = StateSelectTask
		m.SelectedIdx = 0
		m.FocusedTask = nil
		return m, m.cancelActionReminder()
	case "a":
		return m, m.enterNewTaskInput()
	case "s":
		return m, m.enterActionPromptState(StateBreakDone)
	case "q", "ctrl+c", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) handleBreak(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s":
		return m, m.enterActionPromptState(StateBreakDone)
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
		if m.FocusedTask != nil {
			m.LogStore.AppendNew(model.Log{
				TodoID:    m.FocusedTask.ID,
				EventType: model.EventTaskContinue,
			})
		}
		return m, tea.Batch(
			m.cancelActionReminder(),
			tea.Tick(time.Second, func(t time.Time) tea.Msg { return TickMsg(t) }),
		)
	case "d":
		m.completeFocusedTask()
		m.CurrentState = StateSelectTask
		m.SelectedIdx = 0
		m.FocusedTask = nil
		return m, m.cancelActionReminder()
	case "a":
		return m, m.enterNewTaskInput()
	case "n":
		m.CurrentState = StateSelectTask
		m.SelectedIdx = 0
		m.FocusedTask = nil
		return m, m.cancelActionReminder()
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
			if m.FocusedTask != nil {
				m.LogStore.AppendNew(model.Log{
					TodoID:    m.FocusedTask.ID,
					Duration:  model.IntPtr(m.Config.Pomodoro.WorkMinutes),
					EventType: model.EventFocusComplete,
				})
			}
			if m.Config.Pomodoro.Notify {
				return m, tea.Batch(
					alert.Notify("qai", i18n.T("pomo.notify_work_complete")),
					m.enterActionPromptState(StateBreakChoice),
				)
			}
			return m, m.enterActionPromptState(StateBreakChoice)
		case StateBreak:
			if m.Config.Pomodoro.Notify {
				return m, tea.Batch(
					alert.Notify("qai", i18n.T("pomo.notify_break_complete")),
					m.enterActionPromptState(StateBreakDone),
				)
			}
			return m, m.enterActionPromptState(StateBreakDone)
		}
		return m, nil
	}

	return m, tea.Tick(time.Second, func(t time.Time) tea.Msg { return TickMsg(t) })
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
	matchStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
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
	case StateNewTaskInput:
		return m.viewNewTaskInput()
	}
	return ""
}

func (m *Model) viewSelectTask() string {
	var s string
	s += titleStyle.Render(i18n.T("pomo.select_task")) + "\n\n"

	todos := m.getVisibleTodos()
	if len(todos) == 0 && m.SearchQuery != "" {
		s += subtleStyle.Render(i18n.T("pomo.search_no_match")) + "\n"
	}
	for i, t := range todos {
		prefix := "  "
		title := t.Title
		if m.matchesSearch(title) {
			title = highlightMatch(title, m.SearchQuery)
		}
		line := fmt.Sprintf("%s[%d] (P%d) [%s] %s", prefix, t.ID, t.Priority, t.Status, title)
		if i == m.SelectedIdx {
			prefix = "> "
			line = selectedStyle.Render(fmt.Sprintf("%s[%d] (P%d) [%s] ", prefix, t.ID, t.Priority, t.Status)) + title
		}
		s += line + "\n"
	}

	s += "\n"
	if m.SearchMode {
		s += fmt.Sprintf("/%s\n", m.SearchQuery)
	} else if m.SearchQuery != "" {
		s += subtleStyle.Render(i18n.T("pomo.search_active", m.SearchQuery)) + "\n"
	}
	s += subtleStyle.Render(i18n.T("pomo.select_task_hint"))
	return s
}

func highlightMatch(title, query string) string {
	if query == "" {
		return title
	}
	lowerTitle := strings.ToLower(title)
	lowerQuery := strings.ToLower(query)
	var b strings.Builder
	i := 0
	for i < len(title) {
		idx := strings.Index(lowerTitle[i:], lowerQuery)
		if idx < 0 {
			b.WriteString(title[i:])
			break
		}
		b.WriteString(title[i : i+idx])
		b.WriteString(matchStyle.Render(title[i+idx : i+idx+len(query)]))
		i += idx + len(query)
	}
	return b.String()
}

func (m *Model) viewFocus() string {
	var s string
	if m.FocusedTask == nil {
		return subtleStyle.Render("No task selected")
	}
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

	if !m.CompletedAt.IsZero() {
		s += subtleStyle.Render(fmt.Sprintf("Completed at: %s", m.CompletedAt.Format("15:04"))) + "\n\n"
	}
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

func (m *Model) viewNewTaskInput() string {
	var s string
	s += titleStyle.Render(i18n.T("pomo.new_task_title")) + "\n\n"
	s += fmt.Sprintf("> %s\n\n", m.NewTaskTitle)
	s += subtleStyle.Render(i18n.T("pomo.new_task_hint"))
	return s
}

func formatDuration(d time.Duration) string {
	mins := int(d.Minutes())
	secs := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}
