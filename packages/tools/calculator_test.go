package tools_test

import (
	"context"
	tools2 "github.com/stillmatic/gollum/packages/tools"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculator(t *testing.T) {
	calc := tools2.CalculatorTool{}
	var _ tools2.Tool = &calc
	ctx := context.Background()
	t.Run("simple", func(t *testing.T) {
		calcInput := tools2.CalculatorInput{
			Expression: "1 + 1",
		}
		output, err := calc.Run(ctx, &calcInput)
		assert.NoError(t, err)
		assert.Equal(t, "2", output)
	})

	t.Run("simple with env", func(t *testing.T) {
		env := map[string]interface{}{
			"foo": 1,
			"bar": "baz",
		}

		calcInput := tools2.CalculatorInput{
			Expression:  "foo + foo",
			Environment: env,
		}
		output, err := calc.Run(ctx, calcInput)
		assert.NoError(t, err)
		assert.Equal(t, "2", output)
	})
}
