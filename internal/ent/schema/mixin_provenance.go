package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
)

// externalIDFieldName is the field holding the record's identifier in the external system it was sourced from
const externalIDFieldName = "external_id"

// ProvenanceMixin adds the uniform integration provenance fields to every ingest-capable schema and
// opts the schema into stock ingest generation; adding this mixin is the single declaration required
type ProvenanceMixin struct {
	mixin.Schema

	// SchemaType is the schema that implements the SchemaFuncs interface that is using this mixin
	SchemaType any
	// IncludeExternalID adds the external_id field used as the ingest lookup key
	IncludeExternalID bool
	// IncludeExternalIDIndex adds the unique external_id index scoped to the owner
	IncludeExternalIDIndex bool
	// RequiredExternalID makes the external_id field required instead of optional
	RequiredExternalID bool
	// SearchableExternalID includes the external_id field in search
	SearchableExternalID bool
	// ImmutableExternalID prevents the external_id field from changing after create
	ImmutableExternalID bool
}

// provenanceOption configures the ProvenanceMixin
type provenanceOption func(*ProvenanceMixin)

// newProvenanceMixin returns a ProvenanceMixin for the schema with the given options applied
func newProvenanceMixin(schemaType any, opts ...provenanceOption) ProvenanceMixin {
	p := ProvenanceMixin{
		SchemaType:             schemaType,
		IncludeExternalID:      true,
		IncludeExternalIDIndex: true,
	}

	for _, opt := range opts {
		opt(&p)
	}

	return p
}

// withoutExternalID omits the external_id field and its index, for schemas keyed on another field
func withoutExternalID() provenanceOption {
	return func(p *ProvenanceMixin) {
		p.IncludeExternalID = false
		p.IncludeExternalIDIndex = false
	}
}

// withoutExternalIDIndex keeps the external_id field but omits the default unique index
func withoutExternalIDIndex() provenanceOption {
	return func(p *ProvenanceMixin) {
		p.IncludeExternalIDIndex = false
	}
}

// withRequiredExternalID makes the external_id field required instead of optional
func withRequiredExternalID() provenanceOption {
	return func(p *ProvenanceMixin) {
		p.RequiredExternalID = true
	}
}

// withSearchableExternalID includes the external_id field in search
func withSearchableExternalID() provenanceOption {
	return func(p *ProvenanceMixin) {
		p.SearchableExternalID = true
	}
}

// withImmutableExternalID prevents the external_id field from changing after create
func withImmutableExternalID() provenanceOption {
	return func(p *ProvenanceMixin) {
		p.ImmutableExternalID = true
	}
}

// Annotations of the ProvenanceMixin; schemas declaring their own IntegrationMappingSchema override this
func (ProvenanceMixin) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.IntegrationMappingSchema().StockPersist(),
	}
}

// Fields of the ProvenanceMixin
func (p ProvenanceMixin) Fields() []ent.Field {
	fields := []ent.Field{
		field.String("source_definition_id").
			Comment("canonical id of the integration definition that created or last enriched the record").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entx.IntegrationMappingField().SystemControlled(),
			),
		field.String("source_definition_version").
			Comment("integration definition version recorded when the record was created or last enriched").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entx.IntegrationMappingField().SystemControlled().Volatile(),
			),
		field.String("source_instance_id").
			Comment("stable identifier of the external system instance the record was sourced from").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entx.IntegrationMappingField().SystemControlled(),
			),
		field.String("managed_by").
			Comment("id of the integration installation managing the record, empty when the record is unclaimed").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entx.IntegrationMappingField().SystemControlled(),
			),
		field.String("integration_run_id").
			Comment("id of the integration run that last wrote this record").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entx.IntegrationMappingField().SystemControlled().Volatile(),
			),
	}

	if p.IncludeExternalID {
		fields = append(fields, p.externalIDField())
	}

	return fields
}

// externalIDField returns the external_id field configured by the mixin options
func (p ProvenanceMixin) externalIDField() ent.Field {
	annotations := []schema.Annotation{
		entgql.OrderField(externalIDFieldName),
		entx.IntegrationMappingField().LookupKey(),
	}

	if p.SearchableExternalID {
		annotations = append(annotations, entx.FieldSearchable())
	}

	externalID := field.String(externalIDFieldName).
		Comment("identifier of the record in an external system").
		Annotations(annotations...)

	if p.ImmutableExternalID {
		externalID = externalID.Immutable()
	}

	if p.RequiredExternalID {
		return externalID.NotEmpty()
	}

	return externalID.Optional()
}

// Indexes of the ProvenanceMixin
func (p ProvenanceMixin) Indexes() []ent.Index {
	indexes := []ent.Index{
		index.Fields("source_instance_id"),
		index.Fields("integration_run_id"),
	}

	if p.IncludeExternalIDIndex {
		indexes = append(indexes, index.Fields(externalIDFieldName, ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")))
	}

	return indexes
}

// Edges of the ProvenanceMixin
func (p ProvenanceMixin) Edges() []ent.Edge {
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
