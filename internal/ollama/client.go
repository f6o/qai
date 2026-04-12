package ollama

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/f6o/qai/internal/ai"
	"github.com/f6o/qai/internal/model"
	"github.com/ollama/ollama/api"
)

func init() {
	ai.Register("ollama", func(cfg *ai.ProviderConfig) (ai.Provider, error) {
		return &Client{Host: cfg.OllamaHost, Model: cfg.OllamaModel}, nil
	})
}

// Client implements ai.Provider for Ollama.
type Client struct {
	Host  string
	Model string
}

func (c *Client) Suggest(ctx context.Context, tasks []model.Task) (string, error) {
	base, err := url.Parse(c.Host)
	if err != nil {
		return "", fmt.Errorf("invalid ollama host: %w", err)
	}

	client := api.NewClient(base, http.DefaultClient)

	prompt := ai.BuildPrompt(tasks)
	stream := false
	req := &api.GenerateRequest{
		Model:  c.Model,
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
