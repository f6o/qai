package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/f6o/qai/internal/ai"
	"github.com/f6o/qai/internal/model"
)

func init() {
	ai.Register("bedrock", func(cfg *ai.ProviderConfig) (ai.Provider, error) {
		return &Client{Region: cfg.BedrockRegion, ModelID: cfg.BedrockModelID}, nil
	})
}

// Client implements ai.Provider for Amazon Bedrock.
type Client struct {
	Region  string
	ModelID string
}

func (c *Client) Suggest(ctx context.Context, tasks []model.Task) (string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(c.Region))
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	prompt := ai.BuildPrompt(tasks)

	input := &bedrockruntime.ConverseInput{
		ModelId: &c.ModelID,
		Messages: []types.Message{
			{
				Role: types.ConversationRoleUser,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{Value: prompt},
				},
			},
		},
	}

	output, err := client.Converse(ctx, input)
	if err != nil {
		return "", fmt.Errorf("bedrock converse failed: %w", err)
	}

	if output.Output == nil {
		return "", fmt.Errorf("bedrock returned empty output")
	}

	msg, ok := output.Output.(*types.ConverseOutputMemberMessage)
	if !ok {
		return "", fmt.Errorf("unexpected bedrock output type")
	}

	var result string
	for _, block := range msg.Value.Content {
		if text, ok := block.(*types.ContentBlockMemberText); ok {
			result += text.Value
		}
	}

	return result, nil
}
