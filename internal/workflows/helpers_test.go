package workflows

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
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
	assert.Check(t, err)
	assert.Check(t, is.Equal("alpha", value))

	value, err = StringField(&sample{Name: "alpha"}, "name")
	assert.Check(t, err)
	assert.Check(t, is.Equal("alpha", value))

	value, err = StringField(sample{Name: "alpha"}, "missing")
	assert.Check(t, err)
	assert.Check(t, is.Equal("", value))

	value, err = StringField(sample{Name: "alpha"}, "count")
	assert.Check(t, err)
	assert.Check(t, is.Equal("", value))

	_, err = StringField(123, "name")
	assert.Check(t, is.ErrorIs(err, ErrStringFieldUnmarshal))

	var nilSample *sample
	_, err = StringField(nilSample, "name")
	assert.Check(t, is.ErrorIs(err, ErrStringFieldNil))

	_, err = StringField(make(chan int), "name")
	assert.Check(t, is.ErrorIs(err, ErrStringFieldMarshal))
}

// TestStringSliceField verifies string slice extraction behavior
func TestStringSliceField(t *testing.T) {
	type sample struct {
		// Tags holds the sample tag list
		Tags []string `json:"tags"`
	}

	value, err := StringSliceField(sample{Tags: []string{"a", "b"}}, "tags")
	assert.Check(t, err)
	assert.Check(t, is.DeepEqual([]string{"a", "b"}, value))

	value, err = StringSliceField(sample{Tags: nil}, "tags")
	assert.Check(t, err)
	assert.Check(t, is.Nil(value))

	value, err = StringSliceField(sample{Tags: []string{"a"}}, "missing")
	assert.Check(t, err)
	assert.Check(t, is.Nil(value))

	_, err = StringSliceField(123, "tags")
	assert.Check(t, is.ErrorIs(err, ErrStringFieldUnmarshal))

	_, err = StringSliceField(nil, "tags")
	assert.Check(t, is.ErrorIs(err, ErrStringFieldNil))

	type badSample struct {
		// Tags holds invalid tag values for testing
		Tags []any `json:"tags"`
	}
	_, err = StringSliceField(badSample{Tags: []any{"ok", 123}}, "tags")
	assert.Check(t, is.ErrorIs(err, ErrStringSliceFieldInvalid))
}
