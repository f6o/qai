package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Show summary report",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		tasks, err := ctx.TaskStore.Load()
		if err != nil {
			return err
		}

		logs, err := ctx.LogStore.Load()
		if err != nil {
			return err
		}

		cmd.Println("=== Summary Report ===")
		cmd.Println()

		cmd.Println("Tasks:")
		cmd.Printf("  Total: %d\n", len(tasks))

		ideas := ctx.TaskStore.FilterIdeas(tasks)
		todos := ctx.TaskStore.FilterTodos(tasks)
		var doneCount int
		for _, t := range tasks {
			if t.Status == "done" {
				doneCount++
			}
		}
		cmd.Printf("  Ideas: %d\n", len(ideas))
		cmd.Printf("  Todos: %d\n", len(todos))
		cmd.Printf("  Done:  %d\n", doneCount)
		cmd.Println()

		cmd.Println("Logs:")
		cmd.Printf("  Total sessions: %d\n", len(logs))

		var totalMinutes int
		for _, l := range logs {
			totalMinutes += l.Duration
		}
		cmd.Printf("  Total focus time: %d minutes\n", totalMinutes)
		cmd.Println()

		today := time.Now()
		todayLogs := ctx.LogStore.FilterByDate(logs, today.Year(), int(today.Month()), today.Day())
		cmd.Println("Today:")
		cmd.Printf("  Sessions: %d\n", len(todayLogs))
		var todayMinutes int
		for _, l := range todayLogs {
			todayMinutes += l.Duration
		}
		cmd.Printf("  Focus time: %d minutes\n", todayMinutes)

		if len(todayLogs) > 0 {
			cmd.Println()
			cmd.Println("Today's logs:")
			for _, l := range todayLogs {
				fmt.Printf("  - [%s] Task #%d: %s (%d min)\n", l.LoggedAt.Format("15:04"), l.TodoID, l.Content, l.Duration)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
