package enums

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionValues(t *testing.T) {
	expected := []string{"EDITOR", "VIEWER", "BLOCKED", "CREATOR"}
	values := Permission("").Values()

	assert.Equal(t, len(expected), len(values))

	for i, v := range values {
		assert.Equal(t, expected[i], v)
	}
}

func TestPermissionString(t *testing.T) {
	tests := []struct {
		permission Permission
		expected   string
	}{
		{Editor, "EDITOR"},
		{Viewer, "VIEWER"},
		{Blocked, "BLOCKED"},
		{Creator, "CREATOR"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.permission.String())
	}
}

func TestToPermission(t *testing.T) {
	tests := []struct {
		input    string
		expected *Permission
	}{
		{"EDITOR", &Editor},
		{"VIEWER", &Viewer},
		{"BLOCKED", &Blocked},
		{"CREATOR", &Creator},
		{"unknown", nil},
	}

	for _, test := range tests {
		result := ToPermission(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestPermissionMarshalGQL(t *testing.T) {
	tests := []struct {
		permission Permission
		expected   string
	}{
		{Editor, `"EDITOR"`},
		{Viewer, `"VIEWER"`},
		{Blocked, `"BLOCKED"`},
		{Creator, `"CREATOR"`},
	}

	for _, test := range tests {
		var writer strings.Builder
		test.permission.MarshalGQL(&writer)

		assert.Equal(t, test.expected, writer.String())
	}
}

func TestPermissionUnmarshalGQL(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected Permission
		hasError bool
	}{
		{"EDITOR", Editor, false},
		{"VIEWER", Viewer, false},
		{"BLOCKED", Blocked, false},
		{"CREATOR", Creator, false},
		{123, "", true},
	}

	for _, test := range tests {
		var p Permission
		err := p.UnmarshalGQL(test.input)
		if test.hasError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			assert.Equal(t, test.expected, p)
		}
	}
}
