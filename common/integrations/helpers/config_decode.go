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
	return func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeOf([]string{}) {
			return data, nil
		}

		switch v := data.(type) {
		case []string:
			return stringsFromAny(v), nil
		case []any:
			return stringsFromAny(v), nil
		case string:
			return stringsFromAny(v), nil
		default:
			return data, nil
		}
	}
}
