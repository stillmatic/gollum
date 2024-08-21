package mixedbread

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stillmatic/gollum/packages/llm"
	"io"
	"net/http"
)

const (
	apiURL = "https://api.mixedbread.ai/v1/embeddings"
)

type MixedbreadEmbedder struct {
	APIKey string
}

type mixedbreadRequest struct {
	Input              interface{} `json:"input"`
	Model              string      `json:"model"`
	Prompt             string      `json:"prompt,omitempty"`
	Normalized         *bool       `json:"normalized,omitempty"`
	Dimensions         *int        `json:"dimensions,omitempty"`
	EncodingFormat     string      `json:"encoding_format,omitempty"`
	TruncationStrategy string      `json:"truncation_strategy,omitempty"`
}

type mixedbreadResponse struct {
	Model  string `json:"model"`
	Object string `json:"object"`
	Data   []struct {
		Embedding interface{} `json:"embedding"`
		Index     int         `json:"index"`
		Object    string      `json:"object"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Normalized bool `json:"normalized"`
}

func NewMixedbreadEmbedder(apiKey string) *MixedbreadEmbedder {
	return &MixedbreadEmbedder{APIKey: apiKey}
}

func ptr[T any](x T) *T {
	return &x
}

func (e *MixedbreadEmbedder) GenerateEmbedding(ctx context.Context, req llm.EmbedRequest) (*llm.EmbeddingResponse, error) {
	if len(req.Image) > 0 {
		return nil, fmt.Errorf("image embedding not supported by Mixedbread API")
	}

	mixedReq := mixedbreadRequest{
		Input:      req.Input,
		Model:      req.ModelConfig.ModelName,
		Normalized: ptr(true),
	}
	if req.Prompt != "" {
		mixedReq.Prompt = req.Prompt
	}

	if req.Dimensions != 0 {
		mixedReq.Dimensions = &req.Dimensions
	}

	jsonData, err := json.Marshal(mixedReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+e.APIKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	var mixedResp mixedbreadResponse
	err = json.Unmarshal(body, &mixedResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	embeddings := make([]llm.Embedding, len(mixedResp.Data))
	for i, data := range mixedResp.Data {
		switch v := data.Embedding.(type) {
		case []interface{}:
			values := make([]float32, len(v))
			for j, val := range v {
				if f, ok := val.(float64); ok {
					values[j] = float32(f)
				}
			}
			embeddings[i] = llm.Embedding{Values: values}
		default:
			return nil, fmt.Errorf("unexpected embedding format")
		}
	}

	return &llm.EmbeddingResponse{Data: embeddings}, nil
}
