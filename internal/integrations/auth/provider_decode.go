package auth

import (
	"strings"

	"github.com/go-viper/mapstructure/v2"

	integrationconfig "github.com/theopenlane/core/internal/integrations/config"
)

// DecodeProviderData decodes provider metadata into the target struct without failing on unknown keys.
func DecodeProviderData(config map[string]any, target any) error {
	if target == nil {
		return ErrDecodeProviderDataTargetNil
	}
	if len(config) == 0 {
		return nil
	}

	mapDecoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           target,
		TagName:          "json",
		WeaklyTypedInput: true,
		MatchName:        matchProviderKey,
		DecodeHook:       integrationconfig.DefaultMapstructureDecodeHook(),
	})
	if err != nil {
		return err
	}

	return mapDecoder.Decode(config)
}

// matchProviderKey compares map keys to struct fields after normalization
func matchProviderKey(mapKey, fieldName string) bool {
	return normalizeProviderKey(mapKey) == normalizeProviderKey(fieldName)
}

// normalizeProviderKey normalizes provider metadata keys for comparison
func normalizeProviderKey(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	return value
}
