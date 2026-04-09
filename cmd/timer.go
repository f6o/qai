package cmd

import (
	"github.com/f6o/qai/i18n"
	"github.com/spf13/cobra"
)

var timerCmd = &cobra.Command{
	Use:     "timer",
	Aliases: []string{"pomo"},
	Short:   i18n.T("cmd.timer.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}
		return ctx.RunPomodoro(cmd, 0)
	},
}

func init() {
	rootCmd.AddCommand(timerCmd)
}
