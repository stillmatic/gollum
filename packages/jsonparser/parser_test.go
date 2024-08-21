package jsonparser_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stillmatic/gollum/packages/jsonparser"
	"github.com/stretchr/testify/assert"
)

type employee struct {
	Name string `json:"name" yaml:"name" jsonschema:"required,minLength=1,maxLength=100"`
	Age  int    `json:"age" yaml:"age" jsonschema:"minimum=18,maximum=80,required"`
}

type company struct {
	Name      string     `json:"name" yaml:"name"`
	Employees []employee `json:"employees,omitempty" yaml:"employees"`
}

type bulletList []string

var testCo = company{
	Name: "Acme",
	Employees: []employee{
		{
			Name: "John",
			Age:  30,
		},
		{
			Name: "Jane",
			Age:  25,
		},
	},
}
var badEmployees = []employee{
	{
		Name: "John",
		Age:  0,
	},
	{
		Name: "",
		Age:  25,
	},
}

func TestParsers(t *testing.T) {
	t.Run("JSONParser", func(t *testing.T) {
		jsonParser := jsonparser.NewJSONParserGeneric[company](false)
		input, err := json.Marshal(testCo)
		assert.NoError(t, err)

		actual, err := jsonParser.Parse(context.Background(), input)
		assert.NoError(t, err)
		assert.Equal(t, testCo, actual)

		// test failure
		employeeParser := jsonparser.NewJSONParserGeneric[employee](true)
		input2, err := json.Marshal(badEmployees)
		assert.NoError(t, err)
		_, err = employeeParser.Parse(context.Background(), input2)
		assert.Error(t, err)
	})

	t.Run("testbenchmark", func(t *testing.T) {
		jsonParser := jsonparser.NewJSONParserGeneric[company](true)
		input, err := json.Marshal(testCo)
		assert.NoError(t, err)
		actual, err := jsonParser.Parse(context.Background(), input)
		assert.NoError(t, err)
		assert.Equal(t, testCo.Name, actual.Name)
	})
}

func BenchmarkParser(b *testing.B) {
	b.Run("JSONParser-NoValidate", func(b *testing.B) {
		jsonParser := jsonparser.NewJSONParserGeneric[company](false)
		input, err := json.Marshal(testCo)
		assert.NoError(b, err)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			actual, err := jsonParser.Parse(context.Background(), input)
			assert.NoError(b, err)
			assert.Equal(b, testCo, actual)
		}
	})
	b.Run("JSONParser-Validate", func(b *testing.B) {
		jsonParser := jsonparser.NewJSONParserGeneric[company](true)
		input, err := json.Marshal(testCo)
		assert.NoError(b, err)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			actual, err := jsonParser.Parse(context.Background(), input)
			assert.NoError(b, err)
			assert.Equal(b, testCo, actual)
		}
	})
}
