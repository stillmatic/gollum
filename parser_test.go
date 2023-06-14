package gollum_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stillmatic/gollum"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type employee struct {
	Name string `json:"name" yaml:"name" validate:"required"`
	Age  int    `json:"age" yaml:"age" validate:"gte=18,lte=60"`
}

type company struct {
	Name      string      `json:"name" yaml:"name"`
	Employees []*employee `json:"employees" yaml:"employees"`
}

type bulletList []string

var testCo = company{
	Name: "Acme",
	Employees: []*employee{
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
		jsonParser := gollum.NewJSONParser[company](false)
		input, err := json.Marshal(testCo)
		assert.NoError(t, err)

		actual, err := jsonParser.Parse(context.Background(), string(input))
		assert.NoError(t, err)
		assert.Equal(t, testCo, actual)

		// test failure
		employeeParser := gollum.NewJSONParser[employee](true)
		input2, err := json.Marshal(badEmployees)
		assert.NoError(t, err)
		_, err = employeeParser.Parse(context.Background(), string(input2))
		assert.Error(t, err)

	})
	t.Run("YAMLParser", func(t *testing.T) {
		// test struct
		yamlParser := gollum.NewYamlParser[company](true)
		input, err := yaml.Marshal(testCo)
		assert.NoError(t, err)
		actual, err := yamlParser.Parse(context.Background(), string(input))
		assert.NoError(t, err)
		assert.Equal(t, testCo, actual)

		// Test bullet list
		yamlParserBullet := gollum.NewYamlParser[bulletList](false)
		input2 := `- bullet point 1
- bullet point 2
- bullet point 3`
		expected := []string{"bullet point 1", "bullet point 2", "bullet point 3"}
		actual2, err := yamlParserBullet.Parse(context.Background(), input2)
		assert.NoError(t, err)
		for i, v := range actual2 {
			assert.Equal(t, expected[i], v)
		}
	})

	t.Run("fail validation", func(t *testing.T) {
		yamlParser := gollum.NewYamlParser[employee](false)
		input, err := yaml.Marshal(badEmployees)
		assert.NoError(t, err)
		resp, err := yamlParser.Parse(context.Background(), string(input))
		assert.Error(t, err)
		_ = resp
	})
}

type location struct {
	City      string  `json:"city" yaml:"city" validate:"required"`
	State     string  `json:"state" yaml:"state" validate:"required"`
	Country   string  `json:"country" yaml:"country" validate:"required"`
	Latitude  float64 `json:"latitude" yaml:"latitude" validate:"required"`
	Longitude float64 `json:"longitude" yaml:"longitude" validate:"required"`
}

func BenchmarkParser(b *testing.B) {
	b.Run("JSONParser-NoValidate", func(b *testing.B) {
		jsonParser := gollum.NewJSONParser[company](false)
		input, err := json.Marshal(testCo)
		assert.NoError(b, err)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			actual, err := jsonParser.Parse(context.Background(), string(input))
			assert.NoError(b, err)
			assert.Equal(b, testCo, actual)
		}
	})
	b.Run("JSONParser-Validate", func(b *testing.B) {
		jsonParser := gollum.NewJSONParser[company](true)
		input, err := json.Marshal(testCo)
		assert.NoError(b, err)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			actual, err := jsonParser.Parse(context.Background(), string(input))
			assert.NoError(b, err)
			assert.Equal(b, testCo, actual)
		}
	})
}
