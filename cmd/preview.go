package cmd

import (
	"fmt"
	"time"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/markdown"
	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: i18n.T("cmd.preview.short"),
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

		if len(args) > 0 {
			var id int
			fmt.Sscanf(args[0], "%d", &id)
			task := ctx.TaskStore.FindByID(tasks, id)
			if task == nil {
				cmd.Println(i18n.T("cmd.preview.not_found", id))
				return nil
			}

			if task.Status == model.StatusIdea || task.ParentID != nil {
				cmd.Printf("## %s\n\n", task.Title)
				children := ctx.TaskStore.FilterByParentID(tasks, id)
				for _, child := range children {
					checkbox := "[ ]"
					switch child.Status {
					case model.StatusDone:
						checkbox = "[x]"
					case model.StatusDoing:
						checkbox = "[/]"
					}
					cmd.Printf("- %s %s\n", checkbox, child.Title)
				}
			} else {
				cmd.Printf("Title: %s\n", task.Title)
			}
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
	rootCmd.AddCommand(previewCmd)
}
