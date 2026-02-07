package auth

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

// DecodeProviderData decodes provider metadata into the target struct without failing on unknown keys.
func DecodeProviderData(config map[string]any, target any) error {
	if target == nil {
		return ErrDecodeProviderDataTargetNil
	}
	if len(config) == 0 {
		return nil
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           target,
		TagName:          "json",
		WeaklyTypedInput: true,
		MatchName:        matchProviderKey,
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

func matchProviderKey(mapKey, fieldName string) bool {
	return normalizeProviderKey(mapKey) == normalizeProviderKey(fieldName)
}

func normalizeProviderKey(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	return value
}
