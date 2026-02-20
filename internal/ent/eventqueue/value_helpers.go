package eventqueue

import (
	"fmt"
	"strings"
)

// EnumParser converts a normalized string into an enum pointer value
type EnumParser[T ~string] func(string) *T

// ValueAsString converts arbitrary values into non-empty strings
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

// ParseEnum parses enum-like values through the provided enum parser
// Optional invalid values can be provided to force known sentinel values to be treated as parse failures
func ParseEnum[T ~string](raw any, parser EnumParser[T], invalid ...T) (T, bool) {
	var zero T
	if parser == nil {
		return zero, false
	}

	stringValue, ok := ValueAsString(raw)
	if !ok || strings.TrimSpace(stringValue) == "" {
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
