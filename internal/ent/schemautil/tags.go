package schemautil

import "entgo.io/ent/dialect/sql"

// TagsHasFold checks if the value matches while disregarding case sensitivity
func TagsHasFold(column, value string) *sql.Predicate {
	return sql.P(func(b *sql.Builder) {
		b.WriteString("EXISTS (SELECT 1 FROM jsonb_array_elements_text(").
			Ident(column).
			WriteString(") AS tag_values(value)")

		b.WriteString(" WHERE LOWER(tag_values.value) = LOWER(").
			Arg(value).
			WriteString("))")
	})
}
