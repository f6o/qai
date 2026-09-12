package cmd

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var ideaPromoteJSON bool

var ideaPromoteCmd = &cobra.Command{
	Use:          "promote <ID>",
	Short:        i18n.T("cmd.idea_promote.short"),
	SilenceUsage: true,
	Long: `Promote an idea to a todo, non-interactively.
Only the status changes (idea -> todo); ID, priority, and parent are untouched.
Writes a status_change log event. Unknown IDs and non-ideas fail with exit 1 (not idempotent).
Note: promotion does not change priority; a promoted task keeps priority 0 and may be hidden by list --above filters.
Use --json to get exactly one machine-readable line of the updated task record.`,
	Example: `  qai idea promote 42
  qai idea promote --json 42 | jq -r .status`,
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
			return errors.New(i18n.T("cmd.idea_promote.err_not_found", id))
		}
		if task.Status != model.StatusIdea {
			return errors.New(i18n.T("cmd.idea_promote.err_not_idea", id, task.Status))
		}

		task.Status = model.StatusTodo
		tasks, err = ctx.TaskStore.Update(tasks, *task)
		if err != nil {
			return err
		}

		ctx.LogStore.AppendNew(model.Log{
			TodoID:     task.ID,
			EventType:  model.EventStatusChange,
			FromStatus: model.StatusIdea,
			ToStatus:   model.StatusTodo,
		})

		if ideaPromoteJSON {
			if err := json.NewEncoder(cmd.OutOrStdout()).Encode(task); err != nil {
				return err
			}
		} else {
			cmd.Println(i18n.T("cmd.idea_promote.success", task.Title, task.ID))
		}
		return nil
	},
}

func init() {
	ideaPromoteCmd.Flags().BoolVar(&ideaPromoteJSON, "json", false, i18n.T("cmd.idea_promote.flag_json"))
	ideaCmd.AddCommand(ideaPromoteCmd)
}
