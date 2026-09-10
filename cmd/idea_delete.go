package cmd

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/model"
	"github.com/f6o/qai/internal/storage"
	"github.com/spf13/cobra"
)

var ideaDeleteJSON bool
var ideaDeleteForce bool

var ideaDeleteCmd = &cobra.Command{
	Use:          "delete <ID>",
	Short:        i18n.T("cmd.idea_delete.short"),
	SilenceUsage: true,
	Long: `Delete an idea by ID, non-interactively. Only tasks with status idea
can be deleted here; refuse otherwise. Deleting an idea that still has
subtasks fails unless --force removes the whole subtree in one save.
Nothing is written to logs: deletions are not log events.
Use --json to get exactly one machine-readable line: an object for a single
removed record, an array for a --force subtree removal.`,
	Example: `  qai idea delete 7
  qai idea delete --force --json 7 | jq 'length'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		tasks, err := ctx.TaskStore.Load()
		if err != nil {
			return err
		}

		idea := ctx.TaskStore.FindByID(tasks, id)
		if idea == nil {
			return errors.New(i18n.T("cmd.idea_delete.err_not_found", id))
		}
		if idea.Status != model.StatusIdea {
			return errors.New(i18n.T("cmd.idea_delete.err_not_idea", id, string(idea.Status)))
		}

		subtreeIDs := collectSubtreeIDs(ctx.TaskStore, tasks, id)
		children := len(subtreeIDs) - 1
		if children > 0 && !ideaDeleteForce {
			return errors.New(i18n.T("cmd.idea_delete.err_has_children", id, children))
		}

		removedSet := make(map[int]bool, len(subtreeIDs))
		for _, sid := range subtreeIDs {
			removedSet[sid] = true
		}

		remaining := tasks
		var records []model.Task
		for _, t := range tasks {
			if !removedSet[t.ID] {
				continue
			}
			var removed *model.Task
			remaining, removed, err = ctx.TaskStore.Remove(remaining, t.ID)
			if err != nil {
				return err
			}
			records = append(records, *removed)
		}

		if err := ctx.TaskStore.Save(remaining); err != nil {
			return err
		}

		if ideaDeleteJSON {
			if len(records) == 1 {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(records[0])
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(records)
		}
		if len(records) == 1 {
			cmd.Println(i18n.T("cmd.idea_delete.success", records[0].Title, records[0].ID))
		} else {
			cmd.Println(i18n.T("cmd.idea_delete.success_cascade", len(records), id))
		}
		return nil
	},
}

func collectSubtreeIDs(store *storage.TaskStorage, tasks []model.Task, root int) []int {
	ids := []int{root}
	seen := map[int]bool{root: true}
	for i := 0; i < len(ids); i++ {
		for _, child := range store.FilterByParentID(tasks, ids[i]) {
			if !seen[child.ID] {
				seen[child.ID] = true
				ids = append(ids, child.ID)
			}
		}
	}
	return ids
}

func init() {
	ideaDeleteCmd.Flags().BoolVar(&ideaDeleteJSON, "json", false, i18n.T("cmd.idea_delete.flag_json"))
	ideaDeleteCmd.Flags().BoolVar(&ideaDeleteForce, "force", false, i18n.T("cmd.idea_delete.flag_force"))
	ideaCmd.AddCommand(ideaDeleteCmd)
}
