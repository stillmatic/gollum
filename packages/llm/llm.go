//go:generate mockgen -source llm.go -destination internal/mocks/llm.go
package llm

import (
	"context"
)

type ProviderType string

const (
	ProviderAnthropic ProviderType = "anthropic"
	ProviderOpenAI    ProviderType = "openai"
	ProviderGroq      ProviderType = "groq"
	ProviderTogether  ProviderType = "together"
	ProviderGoogle    ProviderType = "google"
)

type ModelConfig struct {
	ProviderType ProviderType
	ModelName    string
	BaseURL      string
}

// MessageOptions are options that can be passed to the model for generating a response.
// NB chua: these are the only ones I use, I assume others are useful too...
type MessageOptions struct {
	MaxTokens   int
	Temperature float32
}

type InferRequest struct {
	Message        string
	Image          *[]byte
	Config         ModelConfig
	MessageOptions MessageOptions
}

type StreamDelta struct {
	Text string
	EOF  bool
}

type Responder interface {
	GenerateResponse(ctx context.Context, req InferRequest) (string, error)
	GenerateResponseAsync(ctx context.Context, req InferRequest) (<-chan StreamDelta, error)
}

// ModelConfigStore is a simple in-memory store for model configurations.
type ModelConfigStore struct {
	configs map[string]ModelConfig
}

func NewModelConfigStore() *ModelConfigStore {
	return &ModelConfigStore{
		configs: configs,
	}
}

func NewModelConfigStoreWithConfigs(configs map[string]ModelConfig) *ModelConfigStore {
	return &ModelConfigStore{
		configs: configs,
	}
}

func (m *ModelConfigStore) GetConfig(configName string) (ModelConfig, bool) {
	config, ok := m.configs[configName]
	return config, ok
}

func (m *ModelConfigStore) GetConfigNames() []string {
	var configNames []string
	for k := range m.configs {
		configNames = append(configNames, k)
	}
	return configNames
}

func (m *ModelConfigStore) AddConfig(configName string, config ModelConfig) {
	m.configs[configName] = config
}
