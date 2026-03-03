package cmd

import (
	"fmt"

	"github.com/f6o/qai/i18n"
	"github.com/spf13/cobra"
)

var ideaListCmd = &cobra.Command{
	Use:   "list",
	Short: i18n.T("cmd.idea_list.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		tasks, err := ctx.TaskStore.Load()
		if err != nil {
			return err
		}

		ideas := ctx.TaskStore.FilterIdeas(tasks)
		if len(ideas) == 0 {
			cmd.Println(i18n.T("cmd.idea_list.empty"))
			return nil
		}

		cmd.Println(i18n.T("cmd.idea_list.header"))
		for _, t := range ideas {
			fmt.Printf("  [%d] %s\n", t.ID, t.Title)
		}
		return nil
	},
}

func init() {
	ideaCmd.AddCommand(ideaListCmd)
}
