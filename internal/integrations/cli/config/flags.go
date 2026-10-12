//go:build examples

package config

import (
	"github.com/spf13/pflag"
)

// ConfigKeyAnnotation is the pflag annotation key used to route a flag value
// into a specific (possibly nested) koanf config path instead of using the
// flag's name verbatim
const ConfigKeyAnnotation = "koanf_key"

// SetConfigKey marks a flag so that loadFlags routes it to the given koanf key
func SetConfigKey(fs *pflag.FlagSet, flagName, key string) {
	_ = fs.SetAnnotation(flagName, ConfigKeyAnnotation, []string{key})
}
