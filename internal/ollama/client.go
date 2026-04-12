package ollama

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/f6o/qai/internal/model"
	"github.com/ollama/ollama/api"
)

func Suggest(ctx context.Context, host, modelName string, tasks []model.Task) (string, error) {
	base, err := url.Parse(host)
	if err != nil {
		return "", fmt.Errorf("invalid ollama host: %w", err)
	}

	client := api.NewClient(base, http.DefaultClient)

	prompt := buildPrompt(tasks)
	stream := false
	req := &api.GenerateRequest{
		Model:  modelName,
		Prompt: prompt,
		Stream: &stream,
	}

	var result string
	err = client.Generate(ctx, req, func(resp api.GenerateResponse) error {
		result = resp.Response
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("ollama generate failed: %w", err)
	}

	return result, nil
}

func buildPrompt(tasks []model.Task) string {
	var sb strings.Builder
	sb.WriteString("Here are my recent tasks:\n")
	for _, t := range tasks {
		fmt.Fprintf(&sb, "- [%s] %s (priority: %d)\n", t.Status, t.Title, t.Priority)
	}
	sb.WriteString("\nBased on these tasks, suggest 3-5 new tasks I should consider. Format each as a single line starting with \"- \".")
	return sb.String()
}
