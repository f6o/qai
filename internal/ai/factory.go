package ai

import "fmt"

// ProviderConfig holds the parameters needed to create a provider.
type ProviderConfig struct {
	Name string

	// Ollama
	OllamaHost  string
	OllamaModel string

	// Bedrock
	BedrockRegion  string
	BedrockModelID string
}

// NewProviderFunc is registered per backend.
type NewProviderFunc func(cfg *ProviderConfig) (Provider, error)

var registry = map[string]NewProviderFunc{}

// Register adds a provider factory.
func Register(name string, fn NewProviderFunc) {
	registry[name] = fn
}

// NewProvider creates a provider from config.
func NewProvider(cfg *ProviderConfig) (Provider, error) {
	fn, ok := registry[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("unknown ai provider: %q", cfg.Name)
	}
	return fn(cfg)
}
