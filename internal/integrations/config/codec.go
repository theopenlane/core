package config

import (
	"maps"

	"github.com/go-viper/mapstructure/v2"

	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
)

// DefaultMapstructureDecodeHook composes the decode hooks used by integration metadata and config decoders.
func DefaultMapstructureDecodeHook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.TextUnmarshallerHookFunc(),
	)
}

// MapOptions controls how one JSON-compatible overlay value is merged into a base map.
type MapOptions struct {
	// PruneZero drops zero-valued fields from the overlay before merge.
	PruneZero bool
	// DeepMerge recursively merges nested map values when true.
	DeepMerge bool
}

// JSONValue overlays one JSON-compatible value onto a cloned base map using the supplied options.
func JSONValue(base map[string]any, value any, options MapOptions) (map[string]any, error) {
	overlay, err := jsonx.ToMap(value)
	if err != nil {
		return nil, err
	}

	if options.PruneZero {
		overlay = mapx.PruneMapZeroAny(overlay)
	}

	out := mapx.DeepCloneMapAny(base)
	if out == nil {
		out = map[string]any{}
	}

	if options.DeepMerge {
		return mapx.DeepMergeMapAny(out, overlay), nil
	}

	maps.Copy(out, overlay)

	return out, nil
}
