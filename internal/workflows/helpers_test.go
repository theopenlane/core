package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestStringField verifies string field extraction behavior
func TestStringField(t *testing.T) {
	type sample struct {
		// Name holds the sample name field
		Name string `json:"name"`
		// Count holds the sample count field
		Count int `json:"count"`
	}

	value, err := StringField(sample{Name: "alpha"}, "name")
	assert.NoError(t, err)
	assert.Equal(t, "alpha", value)

	value, err = StringField(&sample{Name: "alpha"}, "name")
	assert.NoError(t, err)
	assert.Equal(t, "alpha", value)

	value, err = StringField(sample{Name: "alpha"}, "missing")
	assert.NoError(t, err)
	assert.Equal(t, "", value)

	value, err = StringField(sample{Name: "alpha"}, "count")
	assert.NoError(t, err)
	assert.Equal(t, "", value)

	_, err = StringField(123, "name")
	assert.ErrorIs(t, err, ErrStringFieldUnmarshal)

	var nilSample *sample
	_, err = StringField(nilSample, "name")
	assert.ErrorIs(t, err, ErrStringFieldNil)

	_, err = StringField(make(chan int), "name")
	assert.ErrorIs(t, err, ErrStringFieldMarshal)
}

// TestStringSliceField verifies string slice extraction behavior
func TestStringSliceField(t *testing.T) {
	type sample struct {
		// Tags holds the sample tag list
		Tags []string `json:"tags"`
	}

	value, err := StringSliceField(sample{Tags: []string{"a", "b"}}, "tags")
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, value)

	value, err = StringSliceField(sample{Tags: nil}, "tags")
	assert.NoError(t, err)
	assert.Nil(t, value)

	value, err = StringSliceField(sample{Tags: []string{"a"}}, "missing")
	assert.NoError(t, err)
	assert.Nil(t, value)

	_, err = StringSliceField(123, "tags")
	assert.ErrorIs(t, err, ErrStringFieldUnmarshal)

	_, err = StringSliceField(nil, "tags")
	assert.ErrorIs(t, err, ErrStringFieldNil)

	type badSample struct {
		// Tags holds invalid tag values for testing
		Tags []any `json:"tags"`
	}
	_, err = StringSliceField(badSample{Tags: []any{"ok", 123}}, "tags")
	assert.ErrorIs(t, err, ErrStringSliceFieldInvalid)
}
