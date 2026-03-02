package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ideaListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all ideas",
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
			cmd.Println("No ideas yet.")
			return nil
		}

		cmd.Println("Ideas:")
		for _, t := range ideas {
			fmt.Printf("  [%d] %s\n", t.ID, t.Title)
		}
		return nil
	},
}

func init() {
	ideaCmd.AddCommand(ideaListCmd)
}
