package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

var todoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all todos",
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
		if len(todos) == 0 {
			cmd.Println("No todos yet.")
			return nil
		}

		sort.Slice(todos, func(i, j int) bool {
			return todos[i].Priority > todos[j].Priority
		})

		cmd.Println("Todos:")
		for _, t := range todos {
			fmt.Printf("  [%d] [%s] %s\n", t.ID, t.Status, t.Title)
		}
		return nil
	},
}

func init() {
	todoCmd.AddCommand(todoListCmd)
}
