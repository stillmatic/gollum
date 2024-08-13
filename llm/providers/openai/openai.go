package openai

import (
	"context"
	"encoding/base64"
	"io"
	"log/slog"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/stillmatic/gollum/llm"
)

type OpenAIProvider struct {
	client *openai.Client
}

func (p *OpenAIProvider) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
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

func (p *OpenAIProvider) GenerateResponseAsync(ctx context.Context, req InferRequest) (<-chan StreamDelta, error) {
	outChan := make(chan StreamDelta)
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
				case outChan <- StreamDelta{
					Text: content}:
				}
			} else {
				outChan <- StreamDelta{
					EOF: true,
				}
			}
		}
	}()

	return outChan, nil
}
