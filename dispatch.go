package gollum

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type Dispatcher[T any] interface {
	Prompt(ctx context.Context, prompt string) (interface{}, error)
}

type DummyDispatcher[T any] struct{}

func NewDummyDispatcher[T any]() *DummyDispatcher[T] {
	return &DummyDispatcher[T]{}
}

func (d *DummyDispatcher[T]) Prompt(ctx context.Context, prompt string) (T, error) {
	var t T
	return t, nil
}

type OpenAIDispatcherConfig struct {
	Model       *string
	Temperature *float32
	MaxTokens   *int
}

// OpenAIDispatcher dispatches to any OpenAI compatible model.
type OpenAIDispatcher[T any] struct {
	*OpenAIDispatcherConfig
	completer ChatCompleter
	fi        openai.FunctionDefinition
	parser    Parser[T]
}

func NewOpenAIDispatcher[T any](name, description string, completer ChatCompleter, cfg *OpenAIDispatcherConfig) *OpenAIDispatcher[T] {
	var t T
	fi := StructToJsonSchema(name, description, t)
	parser := NewJSONParserGeneric[T](true)
	return &OpenAIDispatcher[T]{
		OpenAIDispatcherConfig: cfg,
		completer:              completer,
		fi:                     openai.FunctionDefinition(fi),
		parser:                 parser,
	}
}

func (d *OpenAIDispatcher[T]) Prompt(ctx context.Context, prompt string) (T, error) {
	var output T
	model := openai.GPT3Dot5Turbo0613
	if d.Model != nil {
		model = *d.Model
	}
	temperature := float32(0.0)
	if d.Temperature != nil {
		temperature = *d.Temperature
	}
	maxTokens := 512
	if d.MaxTokens != nil {
		maxTokens = *d.MaxTokens
	}
	resp, err := d.completer.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		// TODO: configure this
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
		},
		Functions:   []openai.FunctionDefinition{d.fi},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	})
	if err != nil {
		return output, err
	}

	output, err = d.parser.Parse(ctx, []byte(resp.Choices[0].Message.FunctionCall.Arguments))
	if err != nil {
		return output, err
	}

	return output, nil
}
