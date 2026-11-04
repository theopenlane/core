package enums

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestPermissionValues(t *testing.T) {
	expected := []string{"EDITOR", "VIEWER", "BLOCKED", "CREATOR"}
	values := Permission("").Values()

	assert.Check(t, is.Equal(len(expected), len(values)))

	for i, v := range values {
		assert.Check(t, is.Equal(expected[i], v))
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
		assert.Check(t, is.Equal(test.expected, test.permission.String()))
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
		assert.Check(t, is.DeepEqual(test.expected, result))
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

		assert.Check(t, is.Equal(test.expected, writer.String()))
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
			assert.Assert(t, err != nil)
		} else {
			assert.NilError(t, err)
			assert.Check(t, is.Equal(test.expected, p))
		}
	}
}
