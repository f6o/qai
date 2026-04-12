package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/f6o/qai/internal/model"
)

// Provider is the interface that AI backends must implement.
type Provider interface {
	Suggest(ctx context.Context, tasks []model.Task) (string, error)
}

// BuildPrompt constructs the prompt shared by all providers.
func BuildPrompt(tasks []model.Task) string {
	var sb strings.Builder
	sb.WriteString("Here are my recent tasks:\n")
	for _, t := range tasks {
		fmt.Fprintf(&sb, "- [%s] %s (priority: %d)\n", t.Status, t.Title, t.Priority)
	}
	sb.WriteString("\nBased on these tasks, suggest 3-5 new tasks I should consider. Format each as a single line starting with \"- \".")
	return sb.String()
}
