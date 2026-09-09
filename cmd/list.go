package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var listMinPriority int
var listJSON bool

var listCmd = &cobra.Command{
	Use:          "list",
	Short:        i18n.T("cmd.list.short"),
	SilenceUsage: true,
	Long: `List all tasks (ideas and todos).
Use --json to get exactly one machine-readable line: {"ideas":[...],"todos":[...]}.`,
	Example: `  qai list
  qai list --json | jq -r '.todos[].id'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		tasks, err := ctx.TaskStore.Load()
		if err != nil {
			return err
		}

		ideas := ctx.TaskStore.FilterIdeas(tasks)
		todos := ctx.TaskStore.FilterTodos(tasks)

		if listMinPriority > 0 {
			ideas = filterByMinPriority(ideas, listMinPriority)
			todos = filterByMinPriority(todos, listMinPriority)
		}

		sort.Slice(todos, func(i, j int) bool {
			return todos[i].Priority > todos[j].Priority
		})

		if listJSON {
			if ideas == nil {
				ideas = []model.Task{}
			}
			if todos == nil {
				todos = []model.Task{}
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(struct {
				Ideas []model.Task `json:"ideas"`
				Todos []model.Task `json:"todos"`
			}{ideas, todos})
		}

		if len(ideas) > 0 {
			cmd.Println(i18n.T("cmd.idea_list.header"))
			for _, t := range ideas {
				fmt.Printf("  [%d] %s\n", t.ID, t.Title)
			}
			cmd.Println("")
		}

		if len(todos) > 0 {
			cmd.Println(i18n.T("cmd.todo_list.header"))
			for _, t := range todos {
				fmt.Printf("  [%d] [%s] %s\n", t.ID, t.Status, t.Title)
			}
		}

		if len(ideas) == 0 && len(todos) == 0 {
			cmd.Println(i18n.T("cmd.list.empty"))
		}

		return nil
	},
}

func filterByMinPriority(tasks []model.Task, minPriority int) []model.Task {
	var result []model.Task
	for _, t := range tasks {
		if t.Priority >= minPriority {
			result = append(result, t)
		}
	}
	return result
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().IntVarP(&listMinPriority, "above", "A", 0, i18n.T("cmd.list.flag.above"))
	listCmd.Flags().BoolVar(&listJSON, "json", false, i18n.T("cmd.list.flag_json"))
}
