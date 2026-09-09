package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/spf13/cobra"
)

var ideaListJSON bool

var ideaListCmd = &cobra.Command{
	Use:          "list",
	Short:        i18n.T("cmd.idea_list.short"),
	SilenceUsage: true,
	Long: `List all ideas.
Use --json to get exactly one machine-readable line: a JSON array of idea records.`,
	Example: `  qai idea list
  qai idea list --json | jq -r '.[].id'`,
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

		if ideaListJSON {
			if ideas == nil {
				ideas = []model.Task{}
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(ideas)
		}

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
	ideaListCmd.Flags().BoolVar(&ideaListJSON, "json", false, i18n.T("cmd.idea_list.flag_json"))
}
