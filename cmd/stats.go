package cmd

import (
	"encoding/json"
	"time"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var statsJSON bool

var statsCmd = &cobra.Command{
	Use:          "stats",
	Short:        i18n.T("cmd.stats.short"),
	SilenceUsage: true,
	Long: `Show task and focus statistics.
Use --json to get exactly one machine-readable line: a JSON object with tasks, focus, and today sections.`,
	Example: `  qai stats
  qai stats --json | jq '.tasks'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var statsData struct {
			Tasks struct {
				Total int `json:"total"`
				Ideas int `json:"ideas"`
				Todos int `json:"todos"`
				Done  int `json:"done"`
			} `json:"tasks"`
			Focus struct {
				Sessions     int `json:"sessions"`
				TotalMinutes int `json:"total_minutes"`
			} `json:"focus"`
			Today struct {
				Date         string      `json:"date"`
				Sessions     int         `json:"sessions"`
				FocusMinutes int         `json:"focus_minutes"`
				Logs         []model.Log `json:"logs"`
			} `json:"today"`
		}

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

		statsData.Tasks.Total = len(tasks)
		ideas := ctx.TaskStore.FilterIdeas(tasks)
		todos := ctx.TaskStore.FilterTodos(tasks)
		var doneCount int
		for _, t := range tasks {
			if t.Status == "done" {
				doneCount++
			}
		}
		statsData.Tasks.Ideas = len(ideas)
		statsData.Tasks.Todos = len(todos)
		statsData.Tasks.Done = doneCount

		focusLogs := ctx.LogStore.FilterByEventType(logs, "focus_complete")
		statsData.Focus.Sessions = len(focusLogs)
		var totalMinutes int
		for _, l := range focusLogs {
			if l.Duration != nil {
				totalMinutes += *l.Duration
			}
		}
		statsData.Focus.TotalMinutes = totalMinutes

		today := time.Now()
		todayLogs := ctx.LogStore.FilterByDate(logs, today.Year(), int(today.Month()), today.Day())
		todayFocusLogs := ctx.LogStore.FilterByEventType(todayLogs, "focus_complete")
		statsData.Today.Date = today.Format("2006-01-02")
		statsData.Today.Sessions = len(todayFocusLogs)
		var todayMinutes int
		for _, l := range todayFocusLogs {
			if l.Duration != nil {
				todayMinutes += *l.Duration
			}
		}
		statsData.Today.FocusMinutes = todayMinutes
		statsData.Today.Logs = normalizeEventTypes(todayLogs)

		if statsJSON {
			if statsData.Today.Logs == nil {
				statsData.Today.Logs = []model.Log{}
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(statsData)
		}

		cmd.Println(i18n.T("cmd.stats.header"))
		cmd.Println()

		cmd.Println(i18n.T("cmd.stats.tasks_header"))
		cmd.Printf("  "+i18n.T("cmd.stats.tasks_total")+"\n", statsData.Tasks.Total)
		cmd.Printf("  "+i18n.T("cmd.stats.ideas")+"\n", statsData.Tasks.Ideas)
		cmd.Printf("  "+i18n.T("cmd.stats.todos")+"\n", statsData.Tasks.Todos)
		cmd.Printf("  "+i18n.T("cmd.stats.done")+"\n", statsData.Tasks.Done)
		cmd.Println()

		cmd.Println(i18n.T("cmd.stats.logs_header"))
		cmd.Printf("  "+i18n.T("cmd.stats.logs_total_sessions")+"\n", statsData.Focus.Sessions)
		cmd.Printf("  "+i18n.T("cmd.stats.logs_total_focus_time")+"\n", statsData.Focus.TotalMinutes)
		cmd.Println()

		cmd.Println(i18n.T("cmd.stats.today_header"))
		cmd.Printf("  "+i18n.T("cmd.stats.today_sessions")+"\n", statsData.Today.Sessions)
		cmd.Printf("  "+i18n.T("cmd.stats.today_focus_time")+"\n", statsData.Today.FocusMinutes)

		if len(statsData.Today.Logs) > 0 {
			cmd.Println()
			cmd.Println(i18n.T("cmd.stats.today_logs"))
			for _, l := range statsData.Today.Logs {
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
	statsCmd.Flags().BoolVar(&statsJSON, "json", false, i18n.T("cmd.stats.flag_json"))
}
