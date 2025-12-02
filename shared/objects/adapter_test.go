package objects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/shared/objects/mocks"
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
		name        string
		id          string
		idExists    bool
		expectID    string
		expectError error
	}{
		{
			name:        "ID exists",
			id:          "test-id-123",
			idExists:    true,
			expectID:    "test-id-123",
			expectError: nil,
		},
		{
			name:        "ID does not exist",
			id:          "",
			idExists:    false,
			expectID:    "",
			expectError: ErrMutationIDNotFound,
		},
		{
			name:        "empty ID but exists flag is true",
			id:          "",
			idExists:    true,
			expectID:    "",
			expectError: nil,
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

			id, err := adapter.ID()

			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectID, id)
			}
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

	mockMutation.EXPECT().ID().Return("mock-id", nil)
	mockMutation.EXPECT().Type().Return("MockType")

	id, err := mockMutation.ID()
	require.NoError(t, err)
	assert.Equal(t, "mock-id", id)

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

		id, err := adapter.ID()
		require.NoError(t, err)
		assert.Equal(t, "string-id", id)

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

		id, err := adapter.ID()
		require.NoError(t, err)
		assert.Equal(t, "42", id)

		typeName := adapter.Type()
		assert.Equal(t, "Integer", typeName)
	})
}
