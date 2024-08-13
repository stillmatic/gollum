package dispatch

import (
	"context"
	"fmt"
	"github.com/stillmatic/gollum"
	"github.com/stillmatic/gollum/packages/jsonparser"
	"strings"
	"text/template"

	"github.com/sashabaranov/go-openai"
)

type Dispatcher[T any] interface {
	// Prompt generates an object of type T from the given prompt.
	Prompt(ctx context.Context, prompt string) (T, error)
	// PromptTemplate generates an object of type T from a given template.
	// The prompt is then a template string that is rendered with the given values.
	PromptTemplate(ctx context.Context, template *template.Template, values interface{}) (T, error)
}

type DummyDispatcher[T any] struct{}

func NewDummyDispatcher[T any]() *DummyDispatcher[T] {
	return &DummyDispatcher[T]{}
}

func (d *DummyDispatcher[T]) Prompt(ctx context.Context, prompt string) (T, error) {
	var t T
	return t, nil
}

func (d *DummyDispatcher[T]) PromptTemplate(ctx context.Context, template *template.Template, values interface{}) (T, error) {
	var t T
	var sb strings.Builder
	err := template.Execute(&sb, values)
	if err != nil {
		return t, fmt.Errorf("error executing template: %w", err)
	}
	return t, nil
}

type OpenAIDispatcherConfig struct {
	Model       *string
	Temperature *float32
	MaxTokens   *int
}

// OpenAIDispatcher dispatches to any OpenAI compatible model.
// For any type T and prompt, it will generate and parse the response into T.
type OpenAIDispatcher[T any] struct {
	*OpenAIDispatcherConfig
	completer    gollum.ChatCompleter
	ti           openai.Tool
	systemPrompt string
	parser       jsonparser.Parser[T]
}

func NewOpenAIDispatcher[T any](name, description, systemPrompt string, completer gollum.ChatCompleter, cfg *OpenAIDispatcherConfig) *OpenAIDispatcher[T] {
	// note: name must not have spaces - valid json
	// we won't check here but the openai client will throw an error
	var t T
	fi := StructToJsonSchema(name, description, t)
	ti := FunctionInputToTool(fi)
	parser := jsonparser.NewJSONParserGeneric[T](true)
	return &OpenAIDispatcher[T]{
		OpenAIDispatcherConfig: cfg,
		completer:              completer,
		ti:                     ti,
		parser:                 parser,
		systemPrompt:           systemPrompt,
	}
}

func (d *OpenAIDispatcher[T]) Prompt(ctx context.Context, prompt string) (T, error) {
	var output T
	model := openai.GPT3Dot5Turbo1106
	temperature := float32(0.0)
	maxTokens := 512
	if d.OpenAIDispatcherConfig != nil {
		if d.Model != nil {
			model = *d.Model
		}
		if d.Temperature != nil {
			temperature = *d.Temperature
		}
		if d.MaxTokens != nil {
			maxTokens = *d.MaxTokens
		}
	}

	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: d.systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Tools: []openai.Tool{d.ti},
		ToolChoice: openai.ToolChoice{
			Type: "function",
			Function: openai.ToolFunction{
				Name: d.ti.Function.Name,
			}},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	resp, err := d.completer.CreateChatCompletion(ctx, req)
	if err != nil {
		return output, err
	}

	toolOutput := resp.Choices[0].Message.ToolCalls[0].Function.Arguments
	output, err = d.parser.Parse(ctx, []byte(toolOutput))
	if err != nil {
		return output, err
	}

	return output, nil
}

// PromptTemplate generates an object of type T from a given template.
// This is mostly a convenience wrapper around Prompt.
func (d *OpenAIDispatcher[T]) PromptTemplate(ctx context.Context, template *template.Template, values interface{}) (T, error) {
	var t T
	var sb strings.Builder
	err := template.Execute(&sb, values)
	if err != nil {
		return t, fmt.Errorf("error executing template: %w", err)
	}
	return d.Prompt(ctx, sb.String())
}
