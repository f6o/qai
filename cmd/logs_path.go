package cmd

import (
	"github.com/spf13/cobra"
)

var logsPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Display the log file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		cmd.Println(ctx.Config.Data.Logfile)
		return nil
	},
}

func init() {
	logsCmd.AddCommand(logsPathCmd)
}
