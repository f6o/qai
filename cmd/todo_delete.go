package cmd

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/f6o/qai/i18n"
	"github.com/spf13/cobra"
)

var todoDeleteJSON bool

var todoDeleteCmd = &cobra.Command{
	Use:          "delete <ID>",
	Short:        i18n.T("cmd.todo_delete.short"),
	SilenceUsage: true,
	Long: `Delete a task by ID, non-interactively. Any status can be deleted.
Tasks that are parents of other tasks are refused (delete the children first).
Nothing is written to logs: deletions are not log events.
Use --json to get exactly one machine-readable line of the removed task record.`,
	Example: `  qai todo delete 42
  qai todo delete --json 42 | jq -r .id`,
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

		if ctx.TaskStore.FindByID(tasks, id) == nil {
			return errors.New(i18n.T("cmd.todo_delete.err_not_found", id))
		}

		children := ctx.TaskStore.FilterByParentID(tasks, id)
		if len(children) > 0 {
			return errors.New(i18n.T("cmd.todo_delete.err_has_children", id, len(children)))
		}

		tasks, removed, err := ctx.TaskStore.Remove(tasks, id)
		if err != nil {
			return err
		}

		if err := ctx.TaskStore.Save(tasks); err != nil {
			return err
		}

		if todoDeleteJSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(removed)
		}
		cmd.Println(i18n.T("cmd.todo_delete.success", removed.Title, removed.ID))
		return nil
	},
}

func init() {
	todoDeleteCmd.Flags().BoolVar(&todoDeleteJSON, "json", false, i18n.T("cmd.todo_delete.flag_json"))
	todoCmd.AddCommand(todoDeleteCmd)
}
