package cmd

import (
	"context"
	"sort"

	"github.com/f6o/qai/i18n"
	"github.com/f6o/qai/internal/ai"
	_ "github.com/f6o/qai/internal/bedrock"
	_ "github.com/f6o/qai/internal/ollama"
	"github.com/spf13/cobra"
)

var agentSuggestTaskCmd = &cobra.Command{
	Use:   "suggest-task",
	Short: i18n.T("cmd.agent_suggest_task.short"),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := NewAppContext()
		if err != nil {
			return err
		}

		tasks, err := ctx.TaskStore.Load()
		if err != nil {
			return err
		}

		if len(tasks) == 0 {
			cmd.Println(i18n.T("cmd.agent_suggest_task.no_tasks"))
			return nil
		}

		n, err := cmd.Flags().GetInt("number")
		if err != nil {
			return err
		}

		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
		})

		if n > len(tasks) {
			n = len(tasks)
		}
		recent := tasks[:n]

		provider, err := ai.NewProvider(&ai.ProviderConfig{
			Name:           ctx.Config.Agent.Provider,
			OllamaHost:     ctx.Config.Ollama.Host,
			OllamaModel:    ctx.Config.Ollama.Model,
			BedrockRegion:  ctx.Config.Bedrock.Region,
			BedrockModelID: ctx.Config.Bedrock.ModelID,
		})
		if err != nil {
			return err
		}

		cmd.Println(i18n.T("cmd.agent_suggest_task.thinking"))

		result, err := provider.Suggest(context.Background(), recent)
		if err != nil {
			return err
		}

		cmd.Println(result)
		return nil
	},
}

func init() {
	agentSuggestTaskCmd.Flags().IntP("number", "n", 10, i18n.T("cmd.agent_suggest_task.flag_number"))
	agentCmd.AddCommand(agentSuggestTaskCmd)
}
