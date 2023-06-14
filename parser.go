package gollum

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var v = validator.New()

// Parser is an interface for parsing strings into structs
// It is threadsafe and can be used concurrently. The underlying validator is threadsafe as well.
type Parser[T any] interface {
	Parse(ctx context.Context, input string) (T, error)
	GetFormatString() string
}

// JSONParser is a parser that parses arbitrary JSON structs
// It is threadsafe and can be used concurrently. The underlying validator is threadsafe as well.
type JSONParser[T any] struct {
	validate bool
	format   string
}

// NewJSONParser returns a new JSONParser
// If validate is true, the parser will validate the struct using the struct tags
// The underlying struct must have the `validate` tag if validation is enabled, otherwise it will return `InvalidValidationError`
func NewJSONParser[T any](validate bool) *JSONParser[T] {
	var t T
	format := ""
	if validate {
		format += "type "
		tt := reflect.TypeOf(t)
		format += tt.Name() + " struct {\n"
		for i := 0; i < tt.NumField(); i++ {
			field := tt.Field(i)
			format += field.Name + "`json:" + field.Tag.Get("json")
			if field.Tag.Get("validate") != "" {
				format += " validate:" + field.Tag.Get("validate") + "`"
			}
			format += "\n"
		}
		format += "}"
	}

	return &JSONParser[T]{
		validate: validate,
		format:   format,
	}
}

func (p *JSONParser[T]) GetFormatString() string {
	return p.format
}

func (p *JSONParser[T]) Parse(ctx context.Context, input string) (T, error) {
	var t T
	err := json.Unmarshal([]byte(input), &t)
	if err != nil {
		return t, err
	}
	if p.validate {
		err = v.Struct(t)
		if err != nil {
			return t, errors.Wrap(err, "failed to validate struct")
		}
	}

	return t, nil
}

// YAMLParser converts strings into YAML structs
// It is threadsafe and can be used concurrently. The underlying validator is threadsafe as well.
type YAMLParser[T any] struct {
	validate bool
	format   string
}

// NewYamlParser returns a new YAMLParser
// If validate is true, the parser will validate the struct using the struct tags
// The underlying struct must have the `validate` tag if validation is enabled, otherwise it will return `InvalidValidationError`
func NewYamlParser[T any](validate bool) *YAMLParser[T] {
	var t T
	format := ""
	if validate {
		format += "type "
		tt := reflect.TypeOf(t)
		format += tt.Name() + " struct {\n"
		for i := 0; i < tt.NumField(); i++ {
			field := tt.Field(i)
			format += field.Name + "`yaml:" + field.Tag.Get("yaml")
			if field.Tag.Get("validate") != "" {
				format += " validate:" + field.Tag.Get("validate") + "`"
			}
			format += "\n"
		}
		format += "}"
	}

	return &YAMLParser[T]{
		validate: validate,
		format:   format,
	}
}

func (p *YAMLParser[T]) GetFormatString() string {
	return p.format
}

func (p *YAMLParser[T]) Parse(ctx context.Context, input string) (T, error) {
	var t T
	err := yaml.Unmarshal([]byte(input), &t)
	if err != nil {
		return t, err
	}
	if p.validate {
		err = v.Struct(t)
		if err != nil {
			return t, errors.Wrap(err, "failed to validate struct")
		}
	}
	return t, nil
}
