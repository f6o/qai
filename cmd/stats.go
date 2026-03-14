package cmd

import (
	"time"

	"github.com/f6o/qai/i18n"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: i18n.T("cmd.stats.short"),
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

		cmd.Println(i18n.T("cmd.stats.header"))
		cmd.Println()

		cmd.Println(i18n.T("cmd.stats.tasks_header"))
		cmd.Printf("  "+i18n.T("cmd.stats.tasks_total")+"\n", len(tasks))

		ideas := ctx.TaskStore.FilterIdeas(tasks)
		todos := ctx.TaskStore.FilterTodos(tasks)
		var doneCount int
		for _, t := range tasks {
			if t.Status == "done" {
				doneCount++
			}
		}
		cmd.Printf("  "+i18n.T("cmd.stats.ideas")+"\n", len(ideas))
		cmd.Printf("  "+i18n.T("cmd.stats.todos")+"\n", len(todos))
		cmd.Printf("  "+i18n.T("cmd.stats.done")+"\n", doneCount)
		cmd.Println()

		focusLogs := ctx.LogStore.FilterByEventType(logs, "focus_complete")
		cmd.Println(i18n.T("cmd.stats.logs_header"))
		cmd.Printf("  "+i18n.T("cmd.stats.logs_total_sessions")+"\n", len(focusLogs))

		var totalMinutes int
		for _, l := range focusLogs {
			if l.Duration != nil {
				totalMinutes += *l.Duration
			}
		}
		cmd.Printf("  "+i18n.T("cmd.stats.logs_total_focus_time")+"\n", totalMinutes)
		cmd.Println()

		today := time.Now()
		todayLogs := ctx.LogStore.FilterByDate(logs, today.Year(), int(today.Month()), today.Day())
		todayFocusLogs := ctx.LogStore.FilterByEventType(todayLogs, "focus_complete")
		cmd.Println(i18n.T("cmd.stats.today_header"))
		cmd.Printf("  "+i18n.T("cmd.stats.today_sessions")+"\n", len(todayFocusLogs))
		var todayMinutes int
		for _, l := range todayFocusLogs {
			if l.Duration != nil {
				todayMinutes += *l.Duration
			}
		}
		cmd.Printf("  "+i18n.T("cmd.stats.today_focus_time")+"\n", todayMinutes)

		if len(todayLogs) > 0 {
			cmd.Println()
			cmd.Println(i18n.T("cmd.stats.today_logs"))
			for _, l := range todayLogs {
				dur := 0
				if l.Duration != nil {
					dur = *l.Duration
				}
				cmd.Printf("  "+i18n.T("cmd.stats.today_log_item")+"\n", l.LoggedAt.Format("15:04"), l.TodoID, l.EffectiveEventType(), dur)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
