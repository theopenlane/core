package helpers

import (
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

// DecodeConfig decodes a config map into a target struct, respecting defaults on the target.
func DecodeConfig(config map[string]any, target any) error {
	if target == nil {
		return ErrDecodeConfigTargetNil
	}
	if len(config) == 0 {
		return nil
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           target,
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
		MatchName:        matchConfigKey,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.TextUnmarshallerHookFunc(),
			stringSliceDecodeHook(),
		),
	})
	if err != nil {
		return err
	}

	return decoder.Decode(config)
}

func matchConfigKey(mapKey, fieldName string) bool {
	return normalizeConfigKey(mapKey) == normalizeConfigKey(fieldName)
}

func normalizeConfigKey(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	return value
}

func stringSliceDecodeHook() mapstructure.DecodeHookFunc {
	trimmedElem := reflect.TypeFor[TrimmedString]()
	lowerElem := reflect.TypeFor[LowerString]()
	upperElem := reflect.TypeFor[UpperString]()

	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if to.Kind() != reflect.Slice || to.Elem().Kind() != reflect.String {
			return data, nil
		}
		if from == to {
			return data, nil
		}

		values := stringSliceInput(data)
		if values == nil {
			return data, nil
		}

		normalizer := strings.TrimSpace
		switch to.Elem() {
		case lowerElem:
			normalizer = func(value string) string {
				return strings.ToLower(strings.TrimSpace(value))
			}
		case upperElem:
			normalizer = func(value string) string {
				return strings.ToUpper(strings.TrimSpace(value))
			}
		case trimmedElem:
			normalizer = strings.TrimSpace
		default:
			normalizer = strings.TrimSpace
		}

		out := reflect.MakeSlice(to, 0, len(values))
		for _, value := range values {
			cleaned := normalizer(value)
			if cleaned == "" {
				continue
			}
			out = reflect.Append(out, reflect.ValueOf(cleaned).Convert(to.Elem()))
		}
		return out.Interface(), nil
	}
}

func stringSliceInput(data any) []string {
	switch v := data.(type) {
	case []string:
		return v
	case []TrimmedString:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, string(item))
		}
		return out
	case []LowerString:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, string(item))
		}
		return out
	case []UpperString:
		out := make([]string, 0, len(v))
		for _, item := range v {
			out = append(out, string(item))
		}
		return out
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			value := StringFromAny(item)
			if value == "" {
				continue
			}
			out = append(out, value)
		}
		return out
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil
		}
		return strings.Split(trimmed, ",")
	default:
		if value := StringFromAny(v); value != "" {
			return []string{value}
		}
		return nil
	}
}
