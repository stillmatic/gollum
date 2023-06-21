package tools

import (
	"context"
	"strconv"

	"github.com/antonmedv/expr"
	"github.com/pkg/errors"
)

type CalculatorInput struct {
	Expression  string                 `json:"expression" jsonschema:"required" jsonschema_description:"mathematical expression to evaluate"`
	Environment map[string]interface{} `json:"environment,omitempty" jsonschema_description:"optional environment variables to use when evaluating the expression"`
}

type CalculatorTool struct{}

func (c *CalculatorTool) Name() string {
	return "calculator"
}

func (c *CalculatorTool) Description() string {
	return "evaluate mathematical expressions"
}

// Run evaluates a mathematical expression and returns it as a string.
func (c *CalculatorTool) Run(ctx context.Context, input interface{}) (interface{}, error) {
	cinput, ok := input.(CalculatorInput)
	if !ok {
		return "", errors.New("invalid input")
	}

	output, err := expr.Eval(cinput.Expression, cinput.Environment)
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
