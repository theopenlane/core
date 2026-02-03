package helpers

import (
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
