package queryplanner_test

import (
	"context"
	"fmt"
	. "github.com/stillmatic/gollum/packages/queryplanner"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
	"github.com/stillmatic/gollum"
	"github.com/stillmatic/gollum/internal/testutil"
	. "github.com/stillmatic/gollum/queryplanner"
	"github.com/stretchr/testify/assert"
)

func TestQueryPlanner(t *testing.T) {
	godotenv.Load()
	baseAPIURL := "https://api.openai.com/v1/chat/completions"
	openAIKey := os.Getenv("OPENAI_API_KEY")
	assert.NotEmpty(t, openAIKey)

	api := testutil.NewTestAPI(baseAPIURL, openAIKey)
	fi := gollum.StructToJsonSchemaGeneric[QueryPlan]("QueryPlan", "Use this to plan a query.")
	question := "What is the difference between populations of Canada and Jason's home country?"

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are a world class query planning algorithm capable of breaking apart questions into its depenencies queries such that the answers can be used to inform the parent question. Do not answer the questions, simply provide correct compute graph with good specific questions to ask and relevant dependencies. Before you call the function, think step by step to get a better understanding the problem.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("Consider: %s\nGenerate the correct query plan.", question),
		},
	}
	f_ := openai.FunctionDefinition(fi)
	chatRequest := openai.ChatCompletionRequest{
		Messages:    messages,
		Model:       "gpt-3.5-turbo-0613",
		Temperature: 0.0,
		Tools: []openai.Tool{
			{
				Type:     "function",
				Function: &f_,
			},
		},
	}
	ctx := context.Background()
	resp, err := api.SendRequest(ctx, chatRequest)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	parser := gollum.NewJSONParserGeneric[QueryPlan](false)
	queryPlan, err := parser.Parse(ctx, []byte(resp.Choices[0].Message.ToolCalls[0].Function.Arguments))
	assert.NoError(t, err)
	assert.NotNil(t, queryPlan)
	t.Log(queryPlan)
	assert.Equal(t, 0, 1)
}
