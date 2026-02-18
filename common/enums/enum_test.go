package enums

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEnum string

const (
	testFoo testEnum = "FOO"
	testBar testEnum = "BAR"
)

var (
	testInvalid = testEnum("INVALID")
	testValues  = []testEnum{testFoo, testBar}
)

func TestMarshalGQL(t *testing.T) {
	tests := []struct {
		name     string
		input    testEnum
		expected string
	}{
		{
			name:     "standard value",
			input:    testFoo,
			expected: `"FOO"`,
		},
		{
			name:     "another value",
			input:    testBar,
			expected: `"BAR"`,
		},
		{
			name:     "empty value",
			input:    "",
			expected: `""`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			marshalGQL(tc.input, &buf)
			assert.Equal(t, tc.expected, buf.String())
		})
	}
}

func TestUnmarshalGQL(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected testEnum
		wantErr  bool
	}{
		{
			name:     "valid string",
			input:    "FOO",
			expected: testFoo,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:    "integer input",
			input:   42,
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name:    "bool input",
			input:   true,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var result testEnum

			err := unmarshalGQL(&result, tc.input)
			if tc.wantErr {
				require.ErrorIs(t, err, ErrInvalidType)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected testEnum
	}{
		{
			name:     "exact match uppercase",
			input:    "FOO",
			expected: testFoo,
		},
		{
			name:     "lowercase input",
			input:    "foo",
			expected: testFoo,
		},
		{
			name:     "mixed case input",
			input:    "bAr",
			expected: testBar,
		},
		{
			name:     "no match returns fallback",
			input:    "UNKNOWN",
			expected: testInvalid,
		},
		{
			name:     "empty input returns fallback",
			input:    "",
			expected: testInvalid,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parse(tc.input, testValues, &testInvalid)
			require.NotNil(t, result)
			assert.Equal(t, tc.expected, *result)
		})
	}
}

func TestParseReturnsSlicePointer(t *testing.T) {
	vals := []testEnum{testFoo, testBar}

	result := parse("FOO", vals, &testInvalid)

	// the returned pointer should reference the slice element, not a new allocation
	assert.Equal(t, &vals[0], result)
}

func TestStringValues(t *testing.T) {
	tests := []struct {
		name     string
		input    []testEnum
		expected []string
	}{
		{
			name:     "multiple values",
			input:    []testEnum{testFoo, testBar},
			expected: []string{"FOO", "BAR"},
		},
		{
			name:     "single value",
			input:    []testEnum{testFoo},
			expected: []string{"FOO"},
		},
		{
			name:     "empty slice",
			input:    []testEnum{},
			expected: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := stringValues(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
