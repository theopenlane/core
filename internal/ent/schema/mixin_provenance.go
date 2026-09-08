package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/entx"
)

// ProvenanceMixin adds the uniform integration provenance fields to every ingest-capable schema and
// opts the schema into stock ingest generation; adding this mixin is the single declaration required
type ProvenanceMixin struct {
	mixin.Schema
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
				entx.IntegrationMappingField().FromIntegration(),
			),
		field.String("source_definition_version").
			Comment("integration definition version recorded when the record was created or last enriched").
			Optional().
			Annotations(
				entx.IntegrationMappingField().FromIntegration(),
			),
		field.String("source_instance_id").
			Comment("stable identifier of the external system instance the record was sourced from").
			Optional().
			Annotations(
				entx.IntegrationMappingField().FromIntegration(),
			),
		field.String("managed_by").
			Comment("virtual subject id of the integration definition managing the record, empty when user controlled").
			Optional().
			Annotations(
				entx.IntegrationMappingField().FromIntegration(),
			),
	}
}

// Indexes of the ProvenanceMixin
func (ProvenanceMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("source_instance_id"),
	}
}
