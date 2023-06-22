package agents

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
	"github.com/stillmatic/gollum"
	"github.com/stillmatic/gollum/tools"
)

type CalcAgent struct {
	tool          tools.CalculatorTool
	env           map[string]interface{}
	llm           gollum.ChatCompleter
	functionInput openai.FunctionDefinition
	parser        gollum.Parser[tools.CalculatorInput]
}

func NewCalcAgent(llm gollum.ChatCompleter) *CalcAgent {
	// might as well pre-compute it
	fi := gollum.StructToJsonSchemaGeneric[tools.CalculatorInput]("calculator", "evaluate mathematical expressions")
	parser := gollum.NewJSONParserGeneric[tools.CalculatorInput](true)
	return &CalcAgent{
		tool:          tools.CalculatorTool{},
		env:           make(map[string]interface{}),
		llm:           llm,
		functionInput: openai.FunctionDefinition(fi),
		parser:        parser,
	}
}

type CalcAgentInput struct {
	Content string `json:"content" jsonschema:"required" jsonschema_description:"Natural language input to the calculator"`
}

func (c *CalcAgent) Name() string {
	return "calcagent"
}

func (c *CalcAgent) Description() string {
	return "convert natural language and evaluate mathematical expressions"
}

type functionCall struct {
	Name string `json:"name"`
}

func (c *CalcAgent) Run(ctx context.Context, input interface{}) (interface{}, error) {
	cinput, ok := input.(CalcAgentInput)
	if !ok {
		return "", errors.New("invalid input")
	}
	// call LLM to convert natural language to mathematical expression
	resp, err := c.llm.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo0613,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Convert the user's natural language input to a mathematical expression and input to a calculator function. Do not use prior knowledge.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: cinput.Content,
			},
		},
		MaxTokens:    128,
		Functions:    []openai.FunctionDefinition{c.functionInput},
		FunctionCall: functionCall{Name: "calculator"},
	})
	if err != nil {
		return "", errors.Wrap(err, "couldn't call the LLM")
	}
	// parse response
	parsed, err := c.parser.Parse(ctx, []byte(resp.Choices[0].Message.FunctionCall.Arguments))
	if err != nil {
		return "", errors.Wrap(err, "couldn't parse response")
	}
	output, err := c.tool.Run(ctx, parsed)
	if err != nil {
		return "", errors.Wrap(err, "couldn't run expression")
	}
	switch t := output.(type) {
	case string:
		return t, nil
	case int:
		return strconv.Itoa(t), nil
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), nil
	default:
		return "", errors.New("invalid output")
	}
}
