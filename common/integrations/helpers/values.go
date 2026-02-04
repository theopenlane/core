package helpers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

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
	case *string:
		if v == nil {
			return ""
		}
		return strings.TrimSpace(*v)
	case *int:
		if v == nil {
			return ""
		}
		return strconv.Itoa(*v)
	case *int32:
		if v == nil {
			return ""
		}
		return strconv.FormatInt(int64(*v), 10)
	case *int64:
		if v == nil {
			return ""
		}
		return strconv.FormatInt(*v, 10)
	case *uint:
		if v == nil {
			return ""
		}
		return strconv.FormatUint(uint64(*v), 10)
	case *uint32:
		if v == nil {
			return ""
		}
		return strconv.FormatUint(uint64(*v), 10)
	case *uint64:
		if v == nil {
			return ""
		}
		return strconv.FormatUint(*v, 10)
	case *float32:
		if v == nil {
			return ""
		}
		return strconv.FormatFloat(float64(*v), 'f', -1, 32)
	case *float64:
		if v == nil {
			return ""
		}
		return strconv.FormatFloat(*v, 'f', -1, 64)
	case *bool:
		if v == nil {
			return ""
		}
		return strconv.FormatBool(*v)
	}
	return strings.TrimSpace(fmt.Sprint(value))
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
