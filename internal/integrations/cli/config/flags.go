//go:build examples

package config

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ConfigKeyAnnotation is the pflag annotation key used to route a flag value
// into a specific (possibly nested) koanf config path instead of using the
// flag's name verbatim
const ConfigKeyAnnotation = "koanf_key"

// Register adds flags to a cobra command based on struct field tags. When
// prefix is non-empty each flag is annotated with ConfigKeyAnnotation set to
// "prefix.flag-name" so posflag load routes it into the nested koanf path.
//
// Supported tags:
//   - flag: flag name (falls back to koanf tag if not set; "-" skips)
//   - short: single character short flag
//   - default: default value
//   - usage: help text
//
// Supported field types: string, bool, int, []string
func Register(cmd *cobra.Command, prefix string, cfg any) {
	registerTo(cmd.Flags(), prefix, cfg)
}

// SetConfigKey marks a flag so that loadFlags routes it to the given koanf key
func SetConfigKey(fs *pflag.FlagSet, flagName, key string) {
	_ = fs.SetAnnotation(flagName, ConfigKeyAnnotation, []string{key})
}

// registerTo adds flags to a pflag.FlagSet based on struct field tags
func registerTo(fs *pflag.FlagSet, prefix string, cfg any) {
	t := reflect.TypeOf(cfg)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	for i := range t.NumField() {
		registerField(fs, prefix, t.Field(i))
	}
}

// registerField adds a single struct field as a flag and annotates it with
// its nested koanf key when a prefix is provided
func registerField(fs *pflag.FlagSet, prefix string, field reflect.StructField) {
	name := field.Tag.Get("flag")
	if name == "-" {
		return
	}

	if name == "" {
		name = field.Tag.Get("koanf")
	}

	if name == "" {
		return
	}

	usage := field.Tag.Get("usage")
	defVal := field.Tag.Get("default")
	short := field.Tag.Get("short")

	switch field.Type.Kind() {
	case reflect.String:
		if short == "" {
			fs.String(name, defVal, usage)
		} else {
			fs.StringP(name, short, defVal, usage)
		}
	case reflect.Bool:
		boolVal := defVal == "true"
		if short == "" {
			fs.Bool(name, boolVal, usage)
		} else {
			fs.BoolP(name, short, boolVal, usage)
		}
	case reflect.Int:
		intVal, _ := strconv.Atoi(defVal)
		if short == "" {
			fs.Int(name, intVal, usage)
		} else {
			fs.IntP(name, short, intVal, usage)
		}
	case reflect.Slice:
		if field.Type.Elem().Kind() != reflect.String {
			return
		}

		var sliceVal []string
		if defVal != "" {
			sliceVal = strings.Split(defVal, ",")
		}

		if short == "" {
			fs.StringSlice(name, sliceVal, usage)
		} else {
			fs.StringSliceP(name, short, sliceVal, usage)
		}
	default:
		return
	}

	if prefix != "" {
		SetConfigKey(fs, name, prefix+"."+name)
	}
}
