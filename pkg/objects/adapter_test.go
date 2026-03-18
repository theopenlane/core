package objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/pkg/objects/mocks"
)

type testMutation struct {
	id       string
	idExists bool
	typeName string
}

func TestNewGenericMutationAdapter(t *testing.T) {
	mutation := &testMutation{
		id:       "test-id",
		idExists: true,
		typeName: "TestType",
	}

	idFunc := func(m *testMutation) (string, bool) {
		return m.id, m.idExists
	}

	typeFunc := func(m *testMutation) string {
		return m.typeName
	}

	adapter := NewGenericMutationAdapter(mutation, idFunc, typeFunc)

	require.NotNil(t, adapter)
	assert.IsType(t, &GenericMutationAdapter[*testMutation]{}, adapter)
}

func TestGenericMutationAdapter_ID(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		idExists     bool
		expectedID   string
		expectedBool bool
	}{
		{
			name:         "ID exists",
			id:           "test-id-123",
			idExists:     true,
			expectedID:   "test-id-123",
			expectedBool: true,
		},
		{
			name:         "ID does not exist",
			id:           "",
			idExists:     false,
			expectedID:   "",
			expectedBool: false,
		},
		{
			name:         "empty ID but exists flag is true",
			id:           "",
			idExists:     true,
			expectedID:   "",
			expectedBool: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mutation := &testMutation{
				id:       tt.id,
				idExists: tt.idExists,
				typeName: "TestType",
			}

			idFunc := func(m *testMutation) (string, bool) {
				return m.id, m.idExists
			}

			typeFunc := func(m *testMutation) string {
				return m.typeName
			}

			adapter := NewGenericMutationAdapter(mutation, idFunc, typeFunc)

			id, exists := adapter.ID()
			assert.Equal(t, tt.expectedID, id)
			assert.Equal(t, tt.expectedBool, exists)
		})
	}
}

func TestGenericMutationAdapter_Type(t *testing.T) {
	tests := []struct {
		name         string
		typeName     string
		expectedType string
	}{
		{
			name:         "basic type",
			typeName:     "User",
			expectedType: "User",
		},
		{
			name:         "compound type",
			typeName:     "UserProfile",
			expectedType: "UserProfile",
		},
		{
			name:         "empty type",
			typeName:     "",
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mutation := &testMutation{
				id:       "test-id",
				idExists: true,
				typeName: tt.typeName,
			}

			idFunc := func(m *testMutation) (string, bool) {
				return m.id, m.idExists
			}

			typeFunc := func(m *testMutation) string {
				return m.typeName
			}

			adapter := NewGenericMutationAdapter(mutation, idFunc, typeFunc)

			result := adapter.Type()
			assert.Equal(t, tt.expectedType, result)
		})
	}
}

func TestGenericMutationAdapter_WithMock(t *testing.T) {
	mockMutation := mocks.NewMockMutation(t)

	mockMutation.EXPECT().ID().Return("mock-id", true)
	mockMutation.EXPECT().Type().Return("MockType")

	id, exists := mockMutation.ID()
	assert.Equal(t, "mock-id", id)
	assert.True(t, exists)

	typeName := mockMutation.Type()
	assert.Equal(t, "MockType", typeName)
}

func TestGenericMutationAdapter_WithDifferentTypes(t *testing.T) {
	t.Run("string mutation", func(t *testing.T) {
		stringMutation := "test-mutation"

		idFunc := func(s string) (string, bool) {
			return "string-id", true
		}

		typeFunc := func(s string) string {
			return "String"
		}

		adapter := NewGenericMutationAdapter(stringMutation, idFunc, typeFunc)

		id, exists := adapter.ID()
		assert.Equal(t, "string-id", id)
		assert.True(t, exists)

		typeName := adapter.Type()
		assert.Equal(t, "String", typeName)
	})

	t.Run("int mutation", func(t *testing.T) {
		intMutation := 42

		idFunc := func(i int) (string, bool) {
			return "42", true
		}

		typeFunc := func(i int) string {
			return "Integer"
		}

		adapter := NewGenericMutationAdapter(intMutation, idFunc, typeFunc)

		id, exists := adapter.ID()
		assert.Equal(t, "42", id)
		assert.True(t, exists)

		typeName := adapter.Type()
		assert.Equal(t, "Integer", typeName)
	})
}
