package google

import (
	"context"
	"github.com/stillmatic/gollum/packages/llm"
	"google.golang.org/api/option"
	"log/slog"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

type Provider struct {
	client *genai.Client
}

func NewGoogleProvider(ctx context.Context, apiKey string) (*Provider, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, errors.Wrap(err, "google client error")
	}

	return &Provider{client: client}, nil
}

func (p *Provider) getModel(req llm.InferRequest) *genai.GenerativeModel {
	model := p.client.GenerativeModel(req.ModelConfig.ModelName)
	model.SetTemperature(req.MessageOptions.Temperature)
	model.SetMaxOutputTokens(int32(req.MessageOptions.MaxTokens))
	// lol...
	model.SafetySettings = []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockNone},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockNone},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockNone},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockNone},
	}
	model.SetCandidateCount(1)
	return model
}

func (p *Provider) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	if len(req.Messages) > 1 {
		return p.generateResponseChat(ctx, req)
	}
	model := p.getModel(req)
	parts := []genai.Part{genai.Text(req.Messages[0].Content)}
	if req.Messages[0].Image != nil && len(*req.Messages[0].Image) > 0 {
		parts = append(parts, genai.ImageData("png", *req.Messages[0].Image))
	}
	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", errors.Wrap(err, "google generate content error")
	}
	respStr := flattenResponse(resp)

	return respStr, nil
}

func (p *Provider) generateResponseChat(ctx context.Context, req llm.InferRequest) (string, error) {
	model := p.getModel(req)
	hist := make([]*genai.Content, 0, len(req.Messages))
	for _, message := range req.Messages[:len(req.Messages)-1] {
		parts := []genai.Part{genai.Text(message.Content)}
		if message.Image != nil && len(*message.Image) > 0 {
			parts = append(parts, genai.ImageData("png", *message.Image))
		}
		hist = append(hist, &genai.Content{
			Parts: parts,
			Role:  message.Role,
		})
	}
	mostRecentMessage := req.Messages[len(req.Messages)-1]

	// NB chua: this might be a bug but Google doesn't seem to accept multiple parts in the same message
	// in the chat API. So can't send text + image if it exists.
	//mostRecentMessagePart := []genai.Part{genai.Text(mostRecentMessage.Content)}
	//if mostRecentMessage.Image != nil && len(*mostRecentMessage.Image) > 0 {
	//	mostRecentMessagePart = append(mostRecentMessagePart, genai.ImageData("png", *mostRecentMessage.Image))
	//}

	cs := model.StartChat()
	resp, err := cs.SendMessage(ctx, genai.Text(mostRecentMessage.Content))
	if err != nil {
		return "", errors.Wrap(err, "google generate content error")
	}
	respStr := flattenResponse(resp)

	return respStr, nil
}

func (p *Provider) GenerateResponseAsync(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	outChan := make(chan llm.StreamDelta)

	go func() {
		defer close(outChan)

		model := p.getModel(req)

		parts := []genai.Part{genai.Text(req.Message)}
		if req.Image != nil && len(*req.Image) > 0 {
			parts = append(parts, genai.ImageData("png", *req.Image))
		}

		iter := model.GenerateContentStream(ctx, parts...)

		for {
			resp, err := iter.Next()
			if errors.Is(err, iterator.Done) {
				outChan <- llm.StreamDelta{
					EOF: true,
				}
				break
			}
			if err != nil {
				slog.Error("error from gemini stream", "err", err, "req", req.Message, "model", req.Config.ModelName)
				return
			}

			content := flattenResponse(resp)
			if content != "" {
				select {
				case <-ctx.Done():
					return
				case outChan <- llm.StreamDelta{Text: content}:
				}
			}
		}
	}()

	return outChan, nil
}

// flattenResponse flattens the response from the Gemini API into a single string.
func flattenResponse(resp *genai.GenerateContentResponse) string {
	var rtn strings.Builder
	for i, part := range resp.Candidates[0].Content.Parts {
		switch part := part.(type) {
		case genai.Text:
			if i > 0 {
				rtn.WriteString(" ")
			}
			rtn.WriteString(string(part))
		}
	}
	return rtn.String()
}
