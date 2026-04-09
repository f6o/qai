package cmd

import (
	"time"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/markdown"
	"github.com/spf13/cobra"
)

var saveFlag bool

var mdCmd = &cobra.Command{
	Use:   "md",
	Short: i18n.T("cmd.md.short"),
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

		if saveFlag {
			filename, err := gen.Save(tasks, time.Now())
			if err != nil {
				return err
			}
			cmd.Println(i18n.T("cmd.md.success", filename))
			return nil
		}

		content, err := gen.Generate(tasks, time.Now())
		if err != nil {
			return err
		}

		cmd.Print(content)
		return nil
	},
}

func init() {
	mdCmd.Flags().BoolVarP(&saveFlag, "save", "s", false, i18n.T("cmd.md.save_flag"))
	rootCmd.AddCommand(mdCmd)
}
