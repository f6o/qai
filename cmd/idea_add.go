package cmd

import (
	"time"

	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var ideaAddCmd = &cobra.Command{
	Use:   "add [content]",
	Short: "Add a new idea",
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
		cmd.Println("Added idea:", task.Title, "(ID:", task.ID, ")")
		return nil
	},
}

func init() {
	ideaCmd.AddCommand(ideaAddCmd)
}
