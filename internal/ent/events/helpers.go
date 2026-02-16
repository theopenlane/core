package events

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// EnumParser converts a normalized string into an enum pointer value.
type EnumParser[T ~string] func(string) *T

// MutationType resolves the mutation schema type from payload metadata with mutation fallback.
func MutationType(payload *MutationPayload) string {
	if payload == nil {
		return ""
	}

	if mutationType := strings.TrimSpace(payload.MutationType); mutationType != "" {
		return mutationType
	}

	if payload.Mutation != nil {
		return strings.TrimSpace(payload.Mutation.Type())
	}

	return ""
}

// ProposedValue returns a proposed field value from mutation metadata.
func ProposedValue(payload *MutationPayload, field string) (any, bool) {
	if payload == nil || payload.ProposedChanges == nil || field == "" {
		return nil, false
	}

	raw, ok := payload.ProposedChanges[field]
	if !ok {
		return nil, false
	}

	return raw, true
}

// ProposedString returns a proposed field value as string when possible.
func ProposedString(payload *MutationPayload, field string) (string, bool) {
	raw, ok := ProposedValue(payload, field)
	if !ok {
		return "", false
	}

	return ValueAsString(raw)
}

// ValueAsString converts arbitrary values into non-empty strings.
func ValueAsString(raw any) (string, bool) {
	switch value := raw.(type) {
	case nil:
		return "", false
	case string:
		return value, true
	case []byte:
		if len(value) == 0 {
			return "", false
		}

		return string(value), true
	case fmt.Stringer:
		stringValue := value.String()
		if strings.TrimSpace(stringValue) == "" {
			return "", false
		}

		return stringValue, true
	default:
		formatted := fmt.Sprint(value)
		if formatted == "" || formatted == "<nil>" {
			return "", false
		}

		return formatted, true
	}
}

// ValueString converts arbitrary values into a string and returns empty string when conversion fails.
func ValueString(raw any) string {
	value, ok := ValueAsString(raw)
	if !ok {
		return ""
	}

	return value
}

// ParseEnum parses enum-like values through the provided enum parser.
// Optional invalid values can be provided to force known sentinel values to be treated as parse failures.
func ParseEnum[T ~string](raw any, parser EnumParser[T], invalid ...T) (T, bool) {
	var zero T
	if parser == nil {
		return zero, false
	}

	stringValue := ValueString(raw)
	if strings.TrimSpace(stringValue) == "" {
		return zero, false
	}

	parsed := parser(strings.TrimSpace(stringValue))
	if parsed == nil {
		return zero, false
	}

	for _, sentinel := range invalid {
		if *parsed == sentinel {
			return zero, false
		}
	}

	return *parsed, true
}

// ParseEnumPtr parses enum-like values through the provided enum parser and returns a pointer.
func ParseEnumPtr[T ~string](raw any, parser EnumParser[T], invalid ...T) *T {
	parsed, ok := ParseEnum(raw, parser, invalid...)
	if !ok {
		return nil
	}

	return &parsed
}

// CloneStringSliceMap deep-copies map values while dropping blank keys.
func CloneStringSliceMap(values map[string][]string) map[string][]string {
	if len(values) == 0 {
		return nil
	}

	filtered := lo.PickBy(values, func(key string, _ []string) bool { return strings.TrimSpace(key) != "" })
	cloned := lo.MapValues(filtered, func(list []string, _ string) []string { return append([]string(nil), list...) })
	if len(cloned) == 0 {
		return nil
	}

	return cloned
}

// CloneAnyMap shallow-copies map values while dropping blank keys.
func CloneAnyMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}

	cloned := lo.PickBy(values, func(key string, _ any) bool { return strings.TrimSpace(key) != "" })
	if len(cloned) == 0 {
		return nil
	}

	return cloned
}
