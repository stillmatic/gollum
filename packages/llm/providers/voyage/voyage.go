package voyage

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
	apiURL = "https://api.voyageai.com/v1/embeddings"
)

type VoyageAIEmbedder struct {
	APIKey string
}

type voyageAIRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type voyageAIResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func NewVoyageAIEmbedder(apiKey string) *VoyageAIEmbedder {
	return &VoyageAIEmbedder{APIKey: apiKey}
}

func (e *VoyageAIEmbedder) GenerateEmbedding(ctx context.Context, req llm.EmbedRequest) (*llm.EmbeddingResponse, error) {
	if len(req.Image) > 0 {
		return nil, fmt.Errorf("image embedding not supported by Voyage AI")
	}

	if req.Dimensions != 0 {
		return nil, fmt.Errorf("custom dimensions not supported by Voyage AI")
	}

	voyageReq := voyageAIRequest{
		Input: req.Input,
		Model: req.ModelConfig.ModelName,
	}

	jsonData, err := json.Marshal(voyageReq)
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

	var voyageResp voyageAIResponse
	err = json.Unmarshal(body, &voyageResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	embeddings := make([]llm.Embedding, len(voyageResp.Data))
	for i, data := range voyageResp.Data {
		embeddings[i] = llm.Embedding{Values: data.Embedding}
	}

	return &llm.EmbeddingResponse{Data: embeddings}, nil
}
