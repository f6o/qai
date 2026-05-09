package alert

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/beeep"
)

type commandType int

const (
	commandNotify commandType = iota
	commandSchedule
	commandCancel
	commandFire
	commandStop
)

type command struct {
	Type        commandType
	ID          string
	Title       string
	Message     string
	Delay       time.Duration
	RepeatEvery time.Duration
}

type Runner struct {
	commands chan command
	done     chan struct{}
}

var (
	defaultMu     sync.RWMutex
	defaultRunner *Runner
)

func StartDefault() *Runner {
	r := &Runner{
		commands: make(chan command, 16),
		done:     make(chan struct{}),
	}

	defaultMu.Lock()
	defaultRunner = r
	defaultMu.Unlock()

	go r.run()
	return r
}

func (r *Runner) Stop() {
	r.commands <- command{Type: commandStop}
	<-r.done

	defaultMu.Lock()
	if defaultRunner == r {
		defaultRunner = nil
	}
	defaultMu.Unlock()
}

func Notify(title, message string) tea.Cmd {
	return send(command{
		Type:    commandNotify,
		Title:   title,
		Message: message,
	})
}

func Schedule(id, title, message string, delay, repeatEvery time.Duration) tea.Cmd {
	return send(command{
		Type:        commandSchedule,
		ID:          id,
		Title:       title,
		Message:     message,
		Delay:       delay,
		RepeatEvery: repeatEvery,
	})
}

func Cancel(id string) tea.Cmd {
	return send(command{
		Type: commandCancel,
		ID:   id,
	})
}

func send(cmd command) tea.Cmd {
	return func() tea.Msg {
		defaultMu.RLock()
		r := defaultRunner
		defaultMu.RUnlock()
		if r == nil {
			return nil
		}

		r.commands <- cmd
		return nil
	}
}

func (r *Runner) run() {
	timers := map[string]*scheduledAlert{}
	defer func() {
		for _, scheduled := range timers {
			scheduled.timer.Stop()
		}
		close(r.done)
	}()

	for cmd := range r.commands {
		switch cmd.Type {
		case commandNotify:
			beeep.Alert(cmd.Title, cmd.Message, "")
		case commandSchedule:
			if scheduled, ok := timers[cmd.ID]; ok {
				scheduled.timer.Stop()
			}
			timers[cmd.ID] = newScheduledAlert(cmd, r.commands)
		case commandCancel:
			if scheduled, ok := timers[cmd.ID]; ok {
				scheduled.timer.Stop()
				delete(timers, cmd.ID)
			}
		case commandFire:
			scheduled, ok := timers[cmd.ID]
			if !ok {
				continue
			}
			delete(timers, cmd.ID)
			beeep.Alert(scheduled.title, scheduled.message, "")
			if scheduled.repeatEvery > 0 {
				timers[cmd.ID] = newScheduledAlert(command{
					Type:        commandSchedule,
					ID:          cmd.ID,
					Title:       scheduled.title,
					Message:     scheduled.message,
					Delay:       scheduled.repeatEvery,
					RepeatEvery: scheduled.repeatEvery,
				}, r.commands)
			}
		case commandStop:
			return
		}
	}
}

type scheduledAlert struct {
	timer       *time.Timer
	title       string
	message     string
	repeatEvery time.Duration
}

func newScheduledAlert(cmd command, commands chan<- command) *scheduledAlert {
	return &scheduledAlert{
		title:       cmd.Title,
		message:     cmd.Message,
		repeatEvery: cmd.RepeatEvery,
		timer: time.AfterFunc(cmd.Delay, func() {
			commands <- command{
				Type: commandFire,
				ID:   cmd.ID,
			}
		}),
	}
}
