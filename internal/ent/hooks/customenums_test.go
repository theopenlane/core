package hooks

import (
	"testing"

	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
	"github.com/stretchr/testify/assert"
)

func TestTableHasColumn(t *testing.T) {
	table := &schema.Table{
		Name: "test_table",
		Columns: []*schema.Column{
			{Name: "id", Type: field.TypeString},
			{Name: "name", Type: field.TypeString},
			{Name: "deleted_at", Type: field.TypeTime},
		},
	}

	t.Run("returns true for existing column", func(t *testing.T) {
		assert.True(t, tableHasColumn(table, "id"))
		assert.True(t, tableHasColumn(table, "name"))
		assert.True(t, tableHasColumn(table, "deleted_at"))
	})

	t.Run("returns false for non-existent column", func(t *testing.T) {
		assert.False(t, tableHasColumn(table, "nonexistent"))
	})
}

func TestTableHasSoftDelete(t *testing.T) {
	t.Run("returns true when deleted_at column exists", func(t *testing.T) {
		table := &schema.Table{
			Name: "soft_delete_table",
			Columns: []*schema.Column{
				{Name: "id", Type: field.TypeString},
				{Name: "deleted_at", Type: field.TypeTime},
			},
		}
		assert.True(t, tableHasSoftDelete(table))
	})

	t.Run("returns false when deleted_at column does not exist", func(t *testing.T) {
		table := &schema.Table{
			Name: "hard_delete_table",
			Columns: []*schema.Column{
				{Name: "id", Type: field.TypeString},
				{Name: "name", Type: field.TypeString},
			},
		}
		assert.False(t, tableHasSoftDelete(table))
	})
}

func TestIsValidEnumField(t *testing.T) {
	t.Run("returns true for known object types with kind field", func(t *testing.T) {
		assert.True(t, IsValidEnumField("risk", "kind"))
		assert.True(t, IsValidEnumField("risk", "")) // defaults to kind
		assert.True(t, IsValidEnumField("control", "kind"))
	})

	t.Run("returns true for object types with custom field", func(t *testing.T) {
		assert.True(t, IsValidEnumField("entity", "relationship_state"))
		assert.True(t, IsValidEnumField("entity", "security_questionnaire_status"))
	})

	t.Run("returns true for global enums", func(t *testing.T) {
		assert.True(t, IsValidEnumField("", "environment"))
		assert.True(t, IsValidEnumField("", "scope"))
	})

	t.Run("handles case variations via snake_case normalization", func(t *testing.T) {
		assert.True(t, IsValidEnumField("Entity", "RelationshipState"))
		assert.True(t, IsValidEnumField("RISK", "KIND"))
	})

	t.Run("returns false for unknown object type", func(t *testing.T) {
		assert.False(t, IsValidEnumField("nonexistent_object_type_xyz", "kind"))
	})

	t.Run("returns false for unknown field on valid object type", func(t *testing.T) {
		assert.False(t, IsValidEnumField("risk", "nonexistent_field"))
	})

	t.Run("returns false for unknown global field", func(t *testing.T) {
		assert.False(t, IsValidEnumField("", "nonexistent_field_xyz"))
	})
}


func TestFindTablesWithColumn(t *testing.T) {
	t.Run("finds tables with environment_id column", func(t *testing.T) {
		tables := findTablesWithColumn("environment_id")
		assert.NotEmpty(t, tables)

		for _, tbl := range tables {
			assert.NotEmpty(t, tbl.name)
		}
	})

	t.Run("returns empty for nonexistent column", func(t *testing.T) {
		tables := findTablesWithColumn("nonexistent_column_xyz_123")
		assert.Empty(t, tables)
	})
}

func TestFindTableWithColumn(t *testing.T) {
	t.Run("finds table with risk_kind_id column", func(t *testing.T) {
		tbl := findTableWithColumn("risk_kind_id")
		assert.NotNil(t, tbl)
		assert.Equal(t, "risks", tbl.name)
		assert.True(t, tbl.hasSoftDelete)
	})

	t.Run("returns nil for nonexistent column", func(t *testing.T) {
		tbl := findTableWithColumn("nonexistent_column_xyz_123")
		assert.Nil(t, tbl)
	})
}
