package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var todoListJSON bool

var todoListCmd = &cobra.Command{
	Use:          "list",
	Short:        i18n.T("cmd.todo_list.short"),
	SilenceUsage: true,
	Long: `List all todos.
Use --json to get exactly one machine-readable line: a JSON array of todo records.`,
	Example: `  qai todo list
  qai todo list --json | jq -r '.[].id'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		tasks, err := ctx.TaskStore.Load()
		if err != nil {
			return err
		}

		todos := ctx.TaskStore.FilterTodos(tasks)

		sort.Slice(todos, func(i, j int) bool {
			return todos[i].Priority > todos[j].Priority
		})

		if todoListJSON {
			if todos == nil {
				todos = []model.Task{}
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(todos)
		}

		if len(todos) == 0 {
			cmd.Println(i18n.T("cmd.todo_list.empty"))
			return nil
		}

		cmd.Println(i18n.T("cmd.todo_list.header"))
		for _, t := range todos {
			fmt.Printf("  [%d] [%s] %s\n", t.ID, t.Status, t.Title)
		}
		return nil
	},
}

func init() {
	todoCmd.AddCommand(todoListCmd)
	todoListCmd.Flags().BoolVar(&todoListJSON, "json", false, i18n.T("cmd.todo_list.flag_json"))
}
