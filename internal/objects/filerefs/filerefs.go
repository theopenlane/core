package filerefs

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
	"github.com/samber/lo"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/migrate"
	"github.com/theopenlane/core/v2/internal/ent/generated/predicate"
)

// referenced matches files with an inbound reference, built from the generated migration schema
// so new edges and foreign keys are covered without a hand written list
var referenced = file.Or(append(ownedKeys(), inboundKeys()...)...)

// ownedKeys matches the edge foreign keys stored on the files table itself
func ownedKeys() []predicate.File {
	return lo.Map(file.ForeignKeys, func(column string, _ int) predicate.File {
		return func(s *sql.Selector) {
			s.Where(sql.NotNull(s.C(column)))
		}
	})
}

// inboundKeys matches the foreign keys on every other table that point at files
func inboundKeys() []predicate.File {
	return lo.FlatMap(migrate.Tables, func(t *schema.Table, _ int) []predicate.File {
		return lo.FilterMap(t.ForeignKeys, func(fk *schema.ForeignKey, _ int) (predicate.File, bool) {
			if fk.RefTable != migrate.FilesTable {
				return nil, false
			}

			return inbound(t.Name, fk.Columns[0].Name), true
		})
	})
}

// inbound matches files whose id appears in the column of the table
func inbound(table, column string) predicate.File {
	return func(s *sql.Selector) {
		ref := sql.Table(table)

		s.Where(sql.In(s.C(file.FieldID), sql.Select(ref.C(column)).From(ref).Where(sql.NotNull(ref.C(column)))))
	}
}

// Referenced returns the ids with an inbound reference; soft deleted parents count because the referencing tables are read directly
func Referenced(ctx context.Context, client *ent.Client, ids []string) (map[string]struct{}, error) {
	found, err := client.File.Query().Where(file.IDIn(ids...), referenced).IDs(ctx)
	if err != nil {
		return nil, err
	}

	return lo.Keyify(found), nil
}
