package helpers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

// ConfigString returns a trimmed string for a config key when possible.
func ConfigString(config map[string]any, key string) string {
	if len(config) == 0 {
		return ""
	}
	value, ok := config[key]
	if !ok {
		return ""
	}
	return StringFromAny(value)
}

// ConfigInt returns an integer config value or the fallback when not present/convertible.
func ConfigInt(config map[string]any, key string, fallback int) int {
	if len(config) == 0 {
		return fallback
	}
	value, ok := config[key]
	if !ok {
		return fallback
	}
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case json.Number:
		if parsed, err := v.Int64(); err == nil {
			return int(parsed)
		}
		if parsed, err := v.Float64(); err == nil {
			return int(parsed)
		}
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return fallback
		}
		if parsed, err := strconv.Atoi(trimmed); err == nil {
			return parsed
		}
		if parsed, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return int(parsed)
		}
	}
	return fallback
}

// ConfigBool returns a boolean config value or the fallback when not present/convertible.
func ConfigBool(config map[string]any, key string, fallback bool) bool {
	if len(config) == 0 {
		return fallback
	}
	value, ok := config[key]
	if !ok {
		return fallback
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return fallback
		}
		if parsed, err := strconv.ParseBool(trimmed); err == nil {
			return parsed
		}
		if parsed, err := strconv.Atoi(trimmed); err == nil {
			return parsed != 0
		}
	case int:
		return v != 0
	case int32:
		return v != 0
	case int64:
		return v != 0
	case float32:
		return v != 0
	case float64:
		return v != 0
	case json.Number:
		if parsed, err := v.Int64(); err == nil {
			return parsed != 0
		}
		if parsed, err := v.Float64(); err == nil {
			return parsed != 0
		}
	}
	return fallback
}

// ConfigStringSlice returns a slice of trimmed strings for a config key.
func ConfigStringSlice(config map[string]any, key string) []string {
	if len(config) == 0 {
		return nil
	}
	value, ok := config[key]
	if !ok {
		return nil
	}
	return stringsFromAny(value)
}

// StringFromAny returns a trimmed string for common scalar inputs.
func StringFromAny(value any) string {
	if lo.IsNil(value) {
		return ""
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case []byte:
		return strings.TrimSpace(string(v))
	case json.Number:
		return strings.TrimSpace(v.String())
	case fmt.Stringer:
		return strings.TrimSpace(v.String())
	}

	rv := reflect.ValueOf(value)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return ""
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(rv.Interface()))
}

func stringsFromAny(value any) []string {
	if lo.IsNil(value) {
		return nil
	}
	switch v := value.(type) {
	case []string:
		return lo.FilterMap(v, func(item string, _ int) (string, bool) {
			cleaned := strings.TrimSpace(item)
			return cleaned, cleaned != ""
		})
	case []any:
		return lo.FilterMap(v, func(item any, _ int) (string, bool) {
			cleaned := StringFromAny(item)
			return cleaned, cleaned != ""
		})
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil
		}
		parts := strings.Split(trimmed, ",")
		return lo.FilterMap(parts, func(part string, _ int) (string, bool) {
			cleaned := strings.TrimSpace(part)
			return cleaned, cleaned != ""
		})
	default:
		if trimmed := StringFromAny(v); trimmed != "" {
			return []string{trimmed}
		}
		return nil
	}
}

func toStringAnyMap(value any) map[string]any {
	switch v := value.(type) {
	case map[string]any:
		return v
	case map[string]string:
		return lo.MapEntries(v, func(key string, item string) (string, any) {
			return key, item
		})
	default:
		return nil
	}
}
