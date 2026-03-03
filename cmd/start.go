package cmd

import (
	"time"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/markdown"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: i18n.T("cmd.start.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		tasks, err := ctx.TaskStore.Load()
		if err != nil {
			return err
		}

		gen := markdown.NewGenerator(ctx.Config.Data.MarkdownDir)
		filename, err := gen.Save(tasks, time.Now())
		if err != nil {
			return err
		}

		cmd.Println(i18n.T("cmd.start.success", filename))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
