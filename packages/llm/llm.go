//go:generate mockgen -source llm.go -destination internal/mocks/llm.go
package llm

import (
	"context"
)

type ProviderType string

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

type InferMessage struct {
	Content string
	Role    string
	Image   *[]byte

	ShouldCache bool
}

type InferRequest struct {
	Messages []InferMessage

	// ModelConfig describes the model to use for generating a response.
	ModelConfig ModelConfig
	// MessageOptions are options that can be passed to the model for generating a response.
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

// CachableResponder is a responder that can cache prompt inputs on the server.
// The implementation differs per provider.
// For example, Deepseek applies caching automatically.
// Gemini requires users to explicitly cache and create new model instances from the cache.
// Anthropic explicitly caches but does not require a particular ID (hash on their end?)
type CachableResponder interface {
	GenerateResponse(ctx context.Context, req InferRequest) (string, error)
	GenerateResponseAsync(ctx context.Context, req InferRequest) (<-chan StreamDelta, error)

	// CacheObject caches the given object with the given key with the provider.
	CacheObject(ctx context.Context, key string, value InferMessage) error
	GetCachedObject(ctx context.Context, key string) (InferMessage, error)
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