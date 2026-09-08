package cmd

import (
	"encoding/json"
	"time"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var ideaAddJSON bool

var ideaAddCmd = &cobra.Command{
	Use:          "add [content]",
	Short:        i18n.T("cmd.idea_add.short"),
	SilenceUsage: true,
	Long: `Add a new idea non-interactively. Content is a required single argument.
Use --json to get exactly one machine-readable line of the task record.
Keys parent_id/started_at are omitted when unset.`,
	Example: `  qai idea add "new feature"
  qai idea add --json "new feature" | jq -r .id`,
	Args: cobra.ExactArgs(1),
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
		if ideaAddJSON {
			if err := json.NewEncoder(cmd.OutOrStdout()).Encode(task); err != nil {
				return err
			}
		} else {
			cmd.Println(i18n.T("cmd.idea_add.success", task.Title, task.ID))
		}
		return nil
	},
}

func init() {
	ideaAddCmd.Flags().BoolVar(&ideaAddJSON, "json", false, i18n.T("cmd.idea_add.flag_json"))
	ideaCmd.AddCommand(ideaAddCmd)
}
