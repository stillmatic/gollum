package agents_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sashabaranov/go-openai"
	"github.com/stillmatic/gollum/agents"
	mock_gollum "github.com/stillmatic/gollum/internal/mocks"
	"github.com/stillmatic/gollum/tools"
	"github.com/stretchr/testify/assert"
)

func TestCalcAgentMocked(t *testing.T) {
	ctrl := gomock.NewController(t)
	llm := mock_gollum.NewMockChatCompleter(ctrl)
	agent := agents.NewCalcAgent(llm)
	assert.NotNil(t, agent)
	ctx := context.Background()
	inp := agents.CalcAgentInput{
		Content: "What's 2 + 2?",
	}
	expectedResp := tools.CalculatorInput{
		Expression: "2 + 2",
	}
	expectedBytes, err := json.Marshal(expectedResp)
	assert.NoError(t, err)
	expectedChatCompletionResp := openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{
				Message: openai.ChatCompletionMessage{
					Role:         openai.ChatMessageRoleAssistant,
					Content:      "",
					FunctionCall: &openai.FunctionCall{Name: "calc", Arguments: string(expectedBytes)},
				},
			},
		},
	}
	llm.EXPECT().CreateChatCompletion(ctx, gomock.Any()).Return(expectedChatCompletionResp, nil)

	resp, err := agent.Run(ctx, inp)
	assert.NoError(t, err)
	assert.Equal(t, "4", resp.(string))
}

func TestCalcAgentReal(t *testing.T) {
	openai_key := os.Getenv("OPENAI_API_KEY")
	assert.NotEmpty(t, openai_key)
	llm := openai.NewClient(openai_key)

	agent := agents.NewCalcAgent(llm)
	assert.NotNil(t, agent)
	ctx := context.Background()
	inp := agents.CalcAgentInput{
		Content: "What's 2 + 2?",
	}

	resp, err := agent.Run(ctx, inp)
	assert.NoError(t, err)
	assert.Equal(t, "4", resp.(string))
}
