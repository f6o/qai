package cmd

import (
	"time"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
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
		return nil
	},
}

func init() {
	todoAddCmd.Flags().IntP("parent", "p", 0, "Parent idea ID")
	todoCmd.AddCommand(todoAddCmd)
}
