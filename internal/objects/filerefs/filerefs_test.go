//go:build test

package filerefs

import (
	"testing"

	"entgo.io/ent/dialect/sql"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/v2/internal/ent/generated/file"
)

func TestReferencedCoversGeneratedKeys(t *testing.T) {
	selector := sql.Dialect("postgres").Select("*").From(sql.Table(file.Table))
	referenced(selector)

	stmt, _ := selector.Query()

	// a foreign key stored on the files table
	assert.Check(t, is.Contains(stmt, `"files"."platform_architecture_diagrams" IS NOT NULL`))
	// an edge table and a field foreign key on another table
	assert.Check(t, is.Contains(stmt, `FROM "platform_files"`))
	assert.Check(t, is.Contains(stmt, `FROM "action_plans"`))
	assert.Check(t, is.Contains(stmt, `FROM "trust_center_docs"`))
}
