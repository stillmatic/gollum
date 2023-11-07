package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

type TestAPI struct {
	baseAPIURL string
	apiKey     string
	client     *http.Client
}

func NewTestAPI(baseAPIURL, apiKey string) *TestAPI {
	return &TestAPI{
		baseAPIURL: baseAPIURL,
		apiKey:     apiKey,
		client:     &http.Client{},
	}
}

func (api *TestAPI) SendRequest(ctx context.Context, chatRequest openai.ChatCompletionRequest) (*openai.ChatCompletionResponse, error) {
	b, err := json.Marshal(chatRequest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", api.baseAPIURL, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+api.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var chatResponse openai.ChatCompletionResponse
	err = json.NewDecoder(resp.Body).Decode(&chatResponse)
	if err != nil {
		return nil, err
	}

	return &chatResponse, nil
}

func GetRandomEmbedding(n int) []float32 {
	vec := make([]float32, n)
	for i := range vec {
		vec[i] = rand.Float32()
	}
	return vec
}

func GetRandomEmbeddingResponse(n int, dim int) openai.EmbeddingResponse {
	data := make([]openai.Embedding, n)
	for i := range data {
		data[i] = openai.Embedding{
			Embedding: GetRandomEmbedding(dim),
		}
	}
	resp := openai.EmbeddingResponse{
		Data: data,
	}
	return resp
}

func GetRandomChatCompletionResponse(n int) openai.ChatCompletionResponse {
	choices := make([]openai.ChatCompletionChoice, n)
	for i := range choices {
		choices[i] = openai.ChatCompletionChoice{
			Message: openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: fmt.Sprintf("test? %d", i),
			},
		}
	}
	return openai.ChatCompletionResponse{
		Choices: choices,
	}
}
