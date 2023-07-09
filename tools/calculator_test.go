package tools_test

import (
	"context"
	"testing"

	"github.com/stillmatic/gollum/tools"
	"github.com/stretchr/testify/assert"
)

func TestCalculator(t *testing.T) {
	calc := tools.CalculatorTool{}
	var _ tools.Tool = &calc
	ctx := context.Background()
	t.Run("simple", func(t *testing.T) {
		calcInput := tools.CalculatorInput{
			Expression: "1 + 1",
		}
		output, err := calc.Run(ctx, calcInput)
		assert.NoError(t, err)
		assert.Equal(t, "2", output)
	})

	t.Run("simple with env", func(t *testing.T) {
		env := map[string]interface{}{
			"foo": 1,
			"bar": "baz",
		}

		calcInput := tools.CalculatorInput{
			Expression:  "foo + foo",
			Environment: env,
		}
		output, err := calc.Run(ctx, calcInput)
		assert.NoError(t, err)
		assert.Equal(t, "2", output)
	})
	t.Run("test reader/writer", func(t *testing.T) {
		calcInput := tools.CalculatorInput{
			Expression: "1 + 1",
		}
		err := calc.Write(ctx, calcInput)
		assert.NoError(t, err)
		output, err := calc.Read(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "2", output)
	})
}
