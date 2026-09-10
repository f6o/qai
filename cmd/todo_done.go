package cmd

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var todoDoneJSON bool

var todoDoneCmd = &cobra.Command{
	Use:          "done <ID>",
	Short:        i18n.T("cmd.todo_done.short"),
	SilenceUsage: true,
	Long: `Mark a todo or doing task as done, non-interactively.
Writes a status_change log event. Unknown IDs, ideas, and tasks
already done fail with exit 1 (not idempotent).
Use --json to get exactly one machine-readable line of the updated task record.`,
	Example: `  qai todo done 42
  qai todo done --json 42 | jq -r .status`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		tasks, err := ctx.TaskStore.Load()
		if err != nil {
			return err
		}

		task := ctx.TaskStore.FindByID(tasks, id)
		if task == nil {
			return errors.New(i18n.T("cmd.todo_done.err_not_found", id))
		}
		if task.Status == model.StatusIdea {
			return errors.New(i18n.T("cmd.todo_done.err_idea", id))
		}
		if task.Status == model.StatusDone {
			return errors.New(i18n.T("cmd.todo_done.err_already_done", id))
		}

		fromStatus := task.Status
		task.Status = model.StatusDone
		tasks, err = ctx.TaskStore.Update(tasks, *task)
		if err != nil {
			return err
		}

		ctx.LogStore.AppendNew(model.Log{
			TodoID:     task.ID,
			EventType:  model.EventStatusChange,
			FromStatus: fromStatus,
			ToStatus:   model.StatusDone,
		})

		if todoDoneJSON {
			if err := json.NewEncoder(cmd.OutOrStdout()).Encode(task); err != nil {
				return err
			}
		} else {
			cmd.Println(i18n.T("cmd.todo_done.success", task.Title, task.ID))
		}
		return nil
	},
}

func init() {
	todoDoneCmd.Flags().BoolVar(&todoDoneJSON, "json", false, i18n.T("cmd.todo_done.flag_json"))
	todoCmd.AddCommand(todoDoneCmd)
}
