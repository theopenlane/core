package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringField(t *testing.T) {
	type sample struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	assert.Equal(t, "alpha", StringField(sample{Name: "alpha"}, "name"))
	assert.Equal(t, "alpha", StringField(&sample{Name: "alpha"}, "name"))
	assert.Equal(t, "", StringField(sample{Name: "alpha"}, "missing"))
	assert.Equal(t, "", StringField(sample{Name: "alpha"}, "count"))
	assert.Equal(t, "", StringField(123, "name"))

	var nilSample *sample
	assert.Equal(t, "", StringField(nilSample, "name"))
}
