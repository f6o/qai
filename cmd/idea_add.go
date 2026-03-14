package cmd

import (
	"time"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var ideaAddCmd = &cobra.Command{
	Use:   "add [content]",
	Short: i18n.T("cmd.idea_add.short"),
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

		task := model.Task{
			Title:     args[0],
			Status:    model.StatusIdea,
			Priority:  0,
			ParentID:  nil,
			CreatedAt: time.Now(),
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
		cmd.Println(i18n.T("cmd.idea_add.success", task.Title, task.ID))
		return nil
	},
}

func init() {
	ideaCmd.AddCommand(ideaAddCmd)
}
