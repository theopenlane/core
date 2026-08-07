package schema

import (
	"fmt"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/index"
	entmixin "entgo.io/ent/schema/mixin"
)

// EdgeIndexMixin adds an index for every edge on a schema that holds its foreign key in a field on the schema's own table
// Without one, postgres has to scan the table to enforce the constraint whenever the referenced row is deleted or re-keyed,
// and traversing the edge in reverse scans it as well
type EdgeIndexMixin struct {
	entmixin.Schema

	// schemaName is the name of the schema the mixin was built for, used to name the indexes
	schemaName string
	// fields are the foreign key fields that need an index
	fields []string
}

// newEdgeIndexMixin builds the mixin from the edges of the given schema, skipping any field
// that the schema already indexes on its own
func newEdgeIndexMixin(s ent.Interface) EdgeIndexMixin {
	m := EdgeIndexMixin{
		schemaName: getName(s),
	}

	covered := coveredFields(s)

	for _, e := range s.Edges() {
		field := e.Descriptor().Field
		if field == "" || covered[field] {
			continue
		}

		m.fields = append(m.fields, field)
	}

	return m
}

// coveredFields returns the fields that already lead a non partial index on the schema, a
// partial index does not count because postgres cannot use one to enforce a foreign key
func coveredFields(s ent.Interface) map[string]bool {
	covered := map[string]bool{}

	for _, idx := range s.Indexes() {
		desc := idx.Descriptor()
		if len(desc.Fields) == 0 {
			continue
		}

		partial := false

		for _, a := range desc.Annotations {
			if ant, ok := a.(entsql.IndexAnnotation); ok && ant.Where != "" {
				partial = true

				break
			}
		}

		if !partial {
			covered[desc.Fields[0]] = true
		}
	}

	return covered
}

// Indexes of the EdgeIndexMixin
func (m EdgeIndexMixin) Indexes() []ent.Index {
	indexes := make([]ent.Index, 0, len(m.fields))

	for _, field := range m.fields {
		// the storage key is set explicitly so the index cannot collide with a partial index
		// the schema declares on the same column
		indexes = append(indexes, index.Fields(field).
			StorageKey(fmt.Sprintf("%s_%s_idx", m.schemaName, field)))
	}

	return indexes
}
