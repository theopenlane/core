package cp

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/samber/lo"
	"github.com/samber/mo"
	"github.com/theopenlane/utils/contextx"
)


// WithValue is a simplified helper that adds any typed value directly to context
func WithValue[T any](ctx context.Context, value T) context.Context {
	return contextx.With(ctx, value)
}

// GetValue is a simplified helper that retrieves any typed value directly from context
func GetValue[T any](ctx context.Context) mo.Option[T] {
	if value, ok := contextx.From[T](ctx); ok {
		return mo.Some(value)
	}

	return mo.None[T]()
}

// GetValueEquals checks if a context value equals the expected value
func GetValueEquals[T comparable](ctx context.Context, expected T) bool {
	return GetValue[T](ctx).OrElse(*new(T)) == expected
}

// StructToCredentials converts struct fields to credentials map
func StructToCredentials[T any](s T) map[string]string {
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	// Handle pointer types using lo utilities
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return make(map[string]string)
		}
		v = v.Elem()
		t = t.Elem()
	}

	// Only works with structs
	if v.Kind() != reflect.Struct {
		return make(map[string]string)
	}

	return lo.Associate(lo.Range(v.NumField()), func(i int) (string, string) {
		field := t.Field(i)
		value := v.Field(i)

		key := lo.SnakeCase(field.Name)
		val := fmt.Sprintf("%v", value.Interface())

		return key, val
	})
}

// StructFromCredentials converts map[string]string to a struct using JSON marshalling
func StructFromCredentials[T any](credentials map[string]string) (T, error) {
	var result T

	jsonMap := make(map[string]any)
	for key, value := range credentials {
		jsonKey := lo.PascalCase(key)
		jsonMap[jsonKey] = value
	}

	// Marshal to JSON then unmarshal to struct
	jsonData, err := json.Marshal(jsonMap)
	if err != nil {
		return result, fmt.Errorf("failed to marshal credentials: %w", err)
	}

	if err := json.Unmarshal(jsonData, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal to struct: %w", err)
	}

	return result, nil
}
