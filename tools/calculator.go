package tools

import (
	"context"
	"strconv"
	"sync"

	"github.com/antonmedv/expr"
	"github.com/pkg/errors"
)

type CalculatorInput struct {
	Expression  string                 `json:"expression" jsonschema:"required" jsonschema_description:"mathematical expression to evaluate"`
	Environment map[string]interface{} `json:"environment,omitempty" jsonschema_description:"optional environment variables to use when evaluating the expression"`
}

type CalculatorTool struct {
	result string
	err    error
	rwmu   sync.RWMutex
}

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

func (c *CalculatorTool) Write(ctx context.Context, input interface{}) error {
	cinput, ok := input.(CalculatorInput)
	if !ok {
		return errors.New("invalid input")
	}
	defer func() {
		c.rwmu.Lock()
		defer c.rwmu.Unlock()
		res, err := c.Run(ctx, cinput)
		if err != nil {
			c.err = err
			return
		}
		c.result = res.(string)
	}()
	return nil
}

func (c *CalculatorTool) Read(ctx context.Context) (interface{}, error) {
	c.rwmu.Lock()
	defer c.rwmu.Unlock()
	if c.err != nil {
		return "", c.err
	}
	return c.result, nil
}
