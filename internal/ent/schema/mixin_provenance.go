package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
)

// ProvenanceMixin adds the uniform integration provenance fields to every ingest-capable schema and
// opts the schema into stock ingest generation; adding this mixin is the single declaration required
type ProvenanceMixin struct {
	mixin.Schema

	// SchemaType is the schema that implements the SchemaFuncs interface that is using this mixin
	SchemaType any
}

// Annotations of the ProvenanceMixin; schemas declaring their own IntegrationMappingSchema override this
func (ProvenanceMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.IntegrationMappingSchema().StockPersist(),
	}
}

// Fields of the ProvenanceMixin
func (ProvenanceMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("source_definition_id").
			Comment("canonical id of the integration definition that created or last enriched the record").
			Optional().
			Annotations(
				entx.IntegrationMappingField().SystemControlled(),
			),
		field.String("source_definition_version").
			Comment("integration definition version recorded when the record was created or last enriched").
			Optional().
			Annotations(
				entx.IntegrationMappingField().SystemControlled().Volatile(),
			),
		field.String("source_instance_id").
			Comment("stable identifier of the external system instance the record was sourced from").
			Optional().
			Annotations(
				entx.IntegrationMappingField().SystemControlled(),
			),
		field.String("managed_by").
			Comment("id of the integration installation managing the record, empty when the record is unclaimed").
			Optional().
			Annotations(
				entx.IntegrationMappingField().SystemControlled(),
			),
		field.String("integration_run_id").
			Comment("id of the integration run that last wrote this record").
			Optional().
			Annotations(
				entx.IntegrationMappingField().SystemControlled().Volatile(),
			),
	}
}

// Indexes of the ProvenanceMixin
func (ProvenanceMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("source_instance_id"),
		index.Fields("integration_run_id"),
	}
}

// Edges of the ProvenanceMixin
func (m ProvenanceMixin) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: m.SchemaType,
			edgeSchema: IntegrationRun{},
			comment:    "integration runs that have written to this record",
			annotations: []schema.Annotation{
				entgql.QueryField(),
				entgql.MultiOrder(),
				accessmap.EdgeViewCheck(Organization{}.Name()),
			},
		}),
	}
}
