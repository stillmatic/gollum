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

// OpenAIDispatcher dispatches to any OpenAI compatible model.
type OpenAIDispatcher[T any] struct {
	completer ChatCompleter
	fi        openai.FunctionDefinition
	parser    Parser[T]
}

func NewOpenAIDispatcher[T any](name, description string, completer ChatCompleter) *OpenAIDispatcher[T] {
	var t T
	fi := StructToJsonSchema(name, description, t)
	parser := NewJSONParserGeneric[T](true)
	return &OpenAIDispatcher[T]{
		completer: completer,
		fi:        openai.FunctionDefinition(fi),
		parser:    parser,
	}
}

func (d *OpenAIDispatcher[T]) Prompt(ctx context.Context, prompt string) (T, error) {
	var output T
	resp, err := d.completer.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		// TODO: configure this
		Model: openai.GPT3Dot5Turbo0613,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
		},
		Functions: []openai.FunctionDefinition{d.fi},
		// TODO: configure this
		Temperature: 0.0,
		MaxTokens:   1000,
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
