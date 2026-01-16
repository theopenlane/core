package workflows

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringField(t *testing.T) {
	type sample struct {
		Name  string
		Count int
	}

	assert.Equal(t, "alpha", StringField(sample{Name: "alpha"}, "Name"))
	assert.Equal(t, "alpha", StringField(&sample{Name: "alpha"}, "Name"))
	assert.Equal(t, "", StringField(sample{Name: "alpha"}, "Missing"))
	assert.Equal(t, "", StringField(sample{Name: "alpha"}, "Count"))
	assert.Equal(t, "", StringField(123, "Name"))

	var nilSample *sample
	assert.Equal(t, "", StringField(nilSample, "Name"))
}
