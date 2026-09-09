package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var logsJSON bool

var logsCmd = &cobra.Command{
	Use:          "logs",
	Short:        i18n.T("cmd.logs.short"),
	SilenceUsage: true,
	Long: `Show focus logs.
Use --json to get exactly one machine-readable line: a JSON array of log records.`,
	Example: `  qai logs
  qai logs --json | jq -r '.[].id'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		logs, err := ctx.LogStore.Load()
		if err != nil {
			return err
		}

		eventTypeFlag, _ := cmd.Flags().GetString("type")
		if eventTypeFlag != "" {
			logs = ctx.LogStore.FilterByEventType(logs, model.EventType(eventTypeFlag))
		}

		if logsJSON {
			return json.NewEncoder(cmd.OutOrStdout()).Encode(normalizeEventTypes(logs))
		}

		if len(logs) == 0 {
			cmd.Println(i18n.T("cmd.logs.empty"))
			return nil
		}

		for _, l := range logs {
			et := l.EffectiveEventType()
			line := fmt.Sprintf("  [%s] #%d %s", l.LoggedAt.Format("2006-01-02 15:04"), l.TodoID, et)
			if l.Duration != nil {
				line += fmt.Sprintf(" (%d min)", *l.Duration)
			}
			if l.Content != "" {
				line += fmt.Sprintf(" - %s", l.Content)
			}
			cmd.Println(line)
		}
		return nil
	},
}

func normalizeEventTypes(logs []model.Log) []model.Log {
	out := make([]model.Log, len(logs))
	for i, l := range logs {
		l.EventType = l.EffectiveEventType()
		out[i] = l
	}
	return out
}

func init() {
	logsCmd.Flags().StringP("type", "t", "", i18n.T("cmd.logs.type_flag"))
	logsCmd.Flags().BoolVar(&logsJSON, "json", false, i18n.T("cmd.logs.flag_json"))
	rootCmd.AddCommand(logsCmd)
}
