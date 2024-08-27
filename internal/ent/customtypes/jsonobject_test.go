package customtypes_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/ent/customtypes"
)

func TestJSONObjectMarshalGQL(t *testing.T) {
	obj := customtypes.JSONObject{
		"key1": "value1",
		"key2": float64(123),
		"key3": true,
	}

	var buf bytes.Buffer

	obj.MarshalGQL(&buf)

	var result customtypes.JSONObject
	err := json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, obj, result)
}

func TestJSONObjectUnmarshalGQL(t *testing.T) {
	obj := map[string]interface{}{
		"key1": "value1",
		"key2": float64(123),
		"key3": true,
	}

	var result customtypes.JSONObject
	err := result.UnmarshalGQL(obj)
	assert.NoError(t, err)

	for k, v := range obj {
		assert.Equal(t, v, result[k])
	}
}
