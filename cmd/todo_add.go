package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/config"
	"github.com/f6o/qai/internal/flock"
	"github.com/f6o/qai/internal/model"
	"github.com/f6o/qai/internal/pomo"
	"github.com/f6o/qai/internal/storage"
	"github.com/spf13/cobra"
)

var todoAddCmd = &cobra.Command{
	Use:   "add [content]",
	Short: i18n.T("cmd.todo_add.short"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		tasks, err := ctx.TaskStore.Load()
		if err != nil {
			return err
		}

		parentID, err := cmd.Flags().GetInt("parent")
		if err != nil {
			return err
		}

		task := model.Task{
			Title:     args[0],
			Status:    model.StatusTodo,
			Priority:  ctx.Config.Task.DefaultPriority,
			ParentID:  nil,
			CreatedAt: time.Now(),
		}

		if parentID > 0 {
			task.ParentID = &parentID
		}

		tasks, err = ctx.TaskStore.Add(tasks, task)
		if err != nil {
			return err
		}

		task = tasks[len(tasks)-1]
		ctx.LogStore.AppendNew(model.Log{
			TodoID:    task.ID,
			Content:   task.Title,
			EventType: model.EventTaskCreate,
		})
		cmd.Println(i18n.T("cmd.todo_add.success", task.Title, task.ID))

		startPomo, err := cmd.Flags().GetBool("start")
		if err != nil {
			return err
		}
		if startPomo {
			return startPomodoro(cmd, ctx.Config, task.ID)
		}

		return nil
	},
}

func startPomodoro(cmd *cobra.Command, cfg *config.Config, taskID int) error {
	lockPath := filepath.Join(filepath.Dir(cfg.Data.Todofile), "timer.lock")
	fl := flock.New(lockPath)
	locked, err := fl.TryLock()
	if err != nil {
		return fmt.Errorf(i18n.T("cmd.timer.error_lock"), err)
	}
	if !locked {
		return errors.New(i18n.T("cmd.timer.error_locked"))
	}
	defer fl.Unlock()

	ts := storage.NewTaskStorage(cfg.Data.Todofile)
	ls := storage.NewLogStorage(cfg.Data.Logfile)

	m := pomo.NewModel(cfg, ts, ls)
	m.AutoStartTaskID = taskID
	p := tea.NewProgram(&m, tea.WithInput(cmd.InOrStdin()), tea.WithOutput(cmd.OutOrStdout()))

	if _, err := p.Run(); err != nil {
		return fmt.Errorf(i18n.T("cmd.timer.error_run"), err)
	}

	return nil
}

func init() {
	todoAddCmd.Flags().IntP("parent", "p", 0, "Parent idea ID")
	todoAddCmd.Flags().BoolP("start", "s", false, i18n.T("cmd.todo_add.flag_start"))
	todoCmd.AddCommand(todoAddCmd)
}
