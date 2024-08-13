package google

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/stillmatic/gollum/llm"
	"google.golang.org/api/iterator"
)

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

type GoogleProvider struct {
	client *genai.Client
}

func (p *GoogleProvider) GenerateResponse(ctx context.Context, req llm.InferRequest) (string, error) {
	model := p.client.GenerativeModel(req.Config.ModelName)
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

	parts := []genai.Part{genai.Text(req.Message)}
	if req.Image != nil && len(*req.Image) > 0 {
		parts = append(parts, genai.ImageData("png", *req.Image))
	}

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", errors.Wrap(err, "google generate content error")
	}
	respStr := flattenResponse(resp)

	return respStr, nil
}

func (p *GoogleProvider) GenerateResponseAsync(ctx context.Context, req llm.InferRequest) (<-chan llm.StreamDelta, error) {
	outChan := make(chan llm.StreamDelta)

	go func() {
		defer close(outChan)

		model := p.client.GenerativeModel(req.Config.ModelName)
		model.SetTemperature(req.MessageOptions.Temperature)
		model.SetMaxOutputTokens(int32(req.MessageOptions.MaxTokens))
		model.SafetySettings = []*genai.SafetySetting{
			{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockNone},
			{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockNone},
			{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockNone},
			{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockNone},
		}
		model.SetCandidateCount(1)

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
