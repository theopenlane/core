package config

import (
	"encoding/json"
	"reflect"

	"github.com/go-viper/mapstructure/v2"

	"github.com/theopenlane/core/pkg/jsonx"
)

// DefaultMapstructureDecodeHook composes the decode hooks used by integration metadata and config decoders.
func DefaultMapstructureDecodeHook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.TextUnmarshallerHookFunc(),
		mapToRawMessageHook(),
	)
}

func mapToRawMessageHook() mapstructure.DecodeHookFuncType {
	rawMessageType := reflect.TypeFor[json.RawMessage]()

	return func(_ reflect.Type, to reflect.Type, data any) (any, error) {
		if to != rawMessageType {
			return data, nil
		}

		raw, err := jsonx.ToRawMessage(data)
		if err != nil {
			return nil, err
		}

		return raw, nil
	}
}
