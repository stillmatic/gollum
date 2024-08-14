package openai

import (
	"context"
	"encoding/base64"
	"github.com/stillmatic/gollum/packages/llm"
	"io"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
)

type Provider struct {
	client *openai.Client
}

func NewOpenAIProvider(apiKey string) *Provider {
	return &Provider{
		client: openai.NewClient(apiKey),
	}
}

func NewGenericProvider(apiKey string, baseURL string) *Provider {
	genericConfig := openai.DefaultConfig(apiKey)
	genericConfig.BaseURL = baseURL
	return &Provider{
		client: openai.NewClientWithConfig(genericConfig),
	}
}

func NewTogetherProvider(apiKey string) *Provider {
	return NewGenericProvider(apiKey, "https://api.together.xyz/v1")
}

func NewGroqProvider(apiKey string) *Provider {
	return NewGenericProvider(apiKey, "https://api.groq.com/openai/v1/")
}

func NewHyperbolicProvider(apiKey string) *Provider {
	return NewGenericProvider(apiKey, "https://api.hyperbolic.xyz/v1")
}

func NewDeepseekProvider(apiKey string) *Provider {
	return NewGenericProvider(apiKey, "https://api.deepseek.com/v1")
}

func (p *Provider) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	msg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Message,
	}
	if req.Image != nil && len(*req.Image) > 0 {
		b64Image := base64.StdEncoding.EncodeToString(*req.Image)
		msg.MultiContent = []openai.ChatMessagePart{
			{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL:    "data:image/png;base64," + b64Image,
					Detail: openai.ImageURLDetailAuto,
				},
			},
			{
				Type: openai.ChatMessagePartTypeText,
				Text: req.Message,
			},
		}
		msg.Content = ""
	}

	oaiReq := openai.ChatCompletionRequest{
		Model:       req.Config.ModelName,
		Messages:    []openai.ChatCompletionMessage{msg},
		MaxTokens:   req.MessageOptions.MaxTokens,
		Temperature: req.MessageOptions.Temperature,
	}

	res, err := p.client.CreateChatCompletion(ctx, oaiReq)
	if err != nil {
		slog.Error("error from openai", "err", err, "req", req.Message, "model", req.Config.ModelName)
		return "", errors.Wrap(err, "openai chat completion error")
	}

	slog.Debug("got response from openai", "model", req.Config.ModelName, "res", res.Choices[0].Message.Content, "req", req.Message)
	return res.Choices[0].Message.Content, nil
}

func (p *Provider) GenerateResponseAsync(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	outChan := make(chan llm.StreamDelta)
	go func() {
		defer close(outChan)
		msg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: req.Message,
		}
		if req.Image != nil && len(*req.Image) > 0 {
			b64Image := base64.StdEncoding.EncodeToString(*req.Image)
			msg.MultiContent = []openai.ChatMessagePart{
				{
					Type: openai.ChatMessagePartTypeImageURL,
					ImageURL: &openai.ChatMessageImageURL{
						URL:    "data:image/png;base64," + b64Image,
						Detail: openai.ImageURLDetailAuto,
					},
				},
				{
					Type: openai.ChatMessagePartTypeText,
					Text: req.Message,
				},
			}
			msg.Content = ""
		}

		oaiReq := openai.ChatCompletionRequest{
			Model:       req.Config.ModelName,
			Messages:    []openai.ChatCompletionMessage{msg},
			MaxTokens:   req.MessageOptions.MaxTokens,
			Temperature: req.MessageOptions.Temperature,
		}

		stream, err := p.client.CreateChatCompletionStream(ctx, oaiReq)
		if err != nil {
			slog.Error("error from openai", "err", err, "req", req.Message, "model", req.Config.ModelName)
			return
		}
		defer stream.Close()

		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return
			}
			slog.Error("error receiving from openai stream", "err", err)
			return
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			if content != "" {
				select {
				case <-ctx.Done():
					return
				case outChan <- llm.StreamDelta{
					Text: content}:
				}
			} else {
				outChan <- llm.StreamDelta{
					EOF: true,
				}
			}
		}
	}()

	return outChan, nil
}
