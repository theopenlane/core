package hooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetRegistry() {
	enumRegistry.mu.Lock()
	defer enumRegistry.mu.Unlock()
	enumRegistry.objectTypes = make(map[string]string)
	enumRegistry.globalEnums = make(map[string][]string)
}

func TestRegisterEnumSchema(t *testing.T) {
	defer resetRegistry()
	resetRegistry()

	t.Run("registers schema for object type validation", func(t *testing.T) {
		RegisterEnumSchema("task", "tasks")
		RegisterEnumSchema("control", "controls")

		assert.True(t, IsValidObjectType("task"))
		assert.True(t, IsValidObjectType("control"))
		assert.False(t, IsValidObjectType("unknown"))
	})

	t.Run("returns table name for object type", func(t *testing.T) {
		RegisterEnumSchema("risk", "risks")

		assert.Equal(t, "risks", GetTableForObjectType("risk"))
		assert.Equal(t, "", GetTableForObjectType("unknown"))
	})
}

func TestRegisterGlobalEnum(t *testing.T) {
	defer resetRegistry()
	resetRegistry()

	testCases := []struct {
		name           string
		fieldName      string
		tableName      string
		expectedTables []string
	}{
		{
			name:           "register first table for field",
			fieldName:      "environment",
			tableName:      "tasks",
			expectedTables: []string{"tasks"},
		},
		{
			name:           "register second table for same field",
			fieldName:      "environment",
			tableName:      "controls",
			expectedTables: []string{"tasks", "controls"},
		},
		{
			name:           "register different field",
			fieldName:      "scope",
			tableName:      "risks",
			expectedTables: []string{"risks"},
		},
		{
			name:           "duplicate registration ignored",
			fieldName:      "environment",
			tableName:      "tasks",
			expectedTables: []string{"tasks", "controls"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			RegisterGlobalEnum(tc.fieldName, tc.tableName)

			tables := GetGlobalEnumTables(tc.fieldName)
			assert.Equal(t, tc.expectedTables, tables)
		})
	}
}

func TestGetGlobalEnumTables(t *testing.T) {
	defer resetRegistry()
	resetRegistry()

	t.Run("returns empty slice for unregistered field", func(t *testing.T) {
		tables := GetGlobalEnumTables("nonexistent")
		assert.Nil(t, tables)
	})

	t.Run("returns registered tables", func(t *testing.T) {
		RegisterGlobalEnum("environment", "platforms")
		RegisterGlobalEnum("environment", "assets")
		RegisterGlobalEnum("environment", "identity_holders")

		tables := GetGlobalEnumTables("environment")
		assert.Equal(t, []string{"platforms", "assets", "identity_holders"}, tables)
	})
}

func TestRegistryConcurrency(t *testing.T) {
	defer resetRegistry()
	resetRegistry()

	done := make(chan bool)

	for i := 0; i < 100; i++ {
		go func(idx int) {
			RegisterGlobalEnum("concurrent_field", "table_"+string(rune('a'+idx%26)))
			RegisterEnumSchema("schema_"+string(rune('a'+idx%26)), "table_"+string(rune('a'+idx%26)))
			_ = GetGlobalEnumTables("concurrent_field")
			_ = IsValidObjectType("schema_a")
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	tables := GetGlobalEnumTables("concurrent_field")
	assert.NotEmpty(t, tables)
}
