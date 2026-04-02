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

		task := model.Task{
			Title:     args[0],
			Status:    model.StatusIdea,
			Priority:  0,
			ParentID:  nil,
			CreatedAt: time.Now(),
		}

		task, err = ctx.Tasks.AddTask(cmd.Context(), task)
		if err != nil {
			return err
		}

		ctx.Logs.AppendLog(cmd.Context(), model.Log{
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
