package models_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/theopenlane/common/models"
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
	assert.Contains(t, err.Error(), "unsupported time format")
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
	assert.Equal(t, `"2024-04-18T14:00:00Z"`, strings.TrimSpace(buf.String()))
}

func TestToDateTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *time.Time
		wantErr  bool
	}{
		{
			name:  "valid date format YYYY-MM-DD",
			input: "2024-04-18",
			expected: func() *time.Time {
				t := time.Date(2024, 4, 18, 0, 0, 0, 0, time.UTC)
				return &t
			}(),
			wantErr: false,
		},
		{
			name:  "valid date format RFC3339",
			input: "2024-04-18T12:30:00Z",
			expected: func() *time.Time {
				t := time.Date(2024, 4, 18, 12, 30, 0, 0, time.UTC)
				return &t
			}(),
			wantErr: false,
		},
		{
			name:     "empty string input",
			input:    "",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "invalid format",
			input:    "18-04-2024",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := models.ToDateTime(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expected == nil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, *tt.expected, time.Time(*result))
				}
			}
		})
	}
}
