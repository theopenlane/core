package models_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/core/pkg/models"
)

func TestDateTime_Scan_Valid(t *testing.T) {
	var d models.DateTime
	input := time.Date(2024, 5, 10, 15, 30, 0, 0, time.UTC)

	err := d.Scan(input)
	assert.NoError(t, err)
	assert.Equal(t, input, time.Time(d))
}

func TestDateTime_Scan_Invalid(t *testing.T) {
	var d models.DateTime
	err := d.Scan("not-a-time")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot scan type string")
}

func TestDateTime_Value(t *testing.T) {
	d := models.DateTime(time.Date(2024, 4, 18, 0, 0, 0, 0, time.UTC))
	val, err := d.Value()
	assert.NoError(t, err)
	assert.Equal(t, time.Time(d), val.(time.Time))
}

func TestDateTime_UnmarshalCSV(t *testing.T) {
	cases := []struct {
		input    string
		expected time.Time
		isErr    bool
	}{
		{"2024-04-18", time.Date(2024, 4, 18, 0, 0, 0, 0, time.UTC), false},
		{"2024-04-18T00:00:00Z", time.Date(2024, 4, 18, 0, 0, 0, 0, time.UTC), false},
		{"", time.Time{}, false},
		{"invalid", time.Time{}, true},
	}

	for _, c := range cases {
		var d models.DateTime
		err := d.UnmarshalCSV(c.input)
		if c.isErr {
			assert.Error(t, err, "input: %q", c.input)
			continue
		}

		assert.NoError(t, err, "input: %q", c.input)
		assert.Equal(t, c.expected, time.Time(d))
	}
}

func TestDateTime_UnmarshalGQL(t *testing.T) {
	cases := []struct {
		input    any
		expected time.Time
		isErr    bool
	}{
		{"2024-04-18", time.Date(2024, 4, 18, 0, 0, 0, 0, time.UTC), false},
		{"2024-04-18T12:34:56Z", time.Date(2024, 4, 18, 12, 34, 56, 0, time.UTC), false},
		{123, time.Time{}, true},
		{"invalid", time.Time{}, true},
	}

	for _, c := range cases {
		var d models.DateTime
		err := d.UnmarshalGQL(c.input)
		if c.isErr {
			assert.Error(t, err, "input: %v", c.input)
			continue
		}

		assert.NoError(t, err, "input: %v", c.input)
		assert.Equal(t, c.expected, time.Time(d))
	}
}

func TestDateTime_MarshalGQL(t *testing.T) {
	d := models.DateTime(time.Date(2024, 4, 18, 14, 0, 0, 0, time.UTC))
	var buf bytes.Buffer
	d.MarshalGQL(&buf)
	assert.Equal(t, `"2024-04-18"`, strings.TrimSpace(buf.String()))
}
