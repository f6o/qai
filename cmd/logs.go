package cmd

import (
	"fmt"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/f6o/qai/internal/service"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: i18n.T("cmd.logs.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		opts := service.LogListOptions{}
		eventTypeFlag, _ := cmd.Flags().GetString("type")
		if eventTypeFlag != "" {
			et := model.EventType(eventTypeFlag)
			opts.EventType = &et
		}

		logs, err := ctx.Logs.ListLogs(cmd.Context(), opts)
		if err != nil {
			return err
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

func init() {
	logsCmd.Flags().StringP("type", "t", "", i18n.T("cmd.logs.type_flag"))
	rootCmd.AddCommand(logsCmd)
}
