package agents_test

import (
	"context"
	"encoding/json"
	"github.com/stillmatic/gollum/packages/agents"
	"github.com/stillmatic/gollum/packages/tools"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
	mock_gollum "github.com/stillmatic/gollum/internal/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
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
					Role:    openai.ChatMessageRoleAssistant,
					Content: "",
					ToolCalls: []openai.ToolCall{
						{Function: openai.FunctionCall{Name: "calc", Arguments: string(expectedBytes)}},
					},
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
