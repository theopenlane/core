package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
)

// SystemMetadata defines OSCAL-centric system metadata scoped to a program
type SystemMetadata struct {
	SchemaFuncs

	ent.Schema
}

// SchemaSystemMetadata is the canonical schema name
const SchemaSystemMetadata = "system_metadata"

// Name returns the schema name
func (SystemMetadata) Name() string {
	return SchemaSystemMetadata
}

// GetType returns the ent type
func (SystemMetadata) GetType() any {
	return SystemMetadata.Type
}

// PluralName returns the plural schema name
func (SystemMetadata) PluralName() string {
	return pluralize.NewClient().Plural(SchemaSystemMetadata)
}

// Fields of the SystemMetadata.
func (SystemMetadata) Fields() []ent.Field {
	return []ent.Field{
		field.String("program_id").
			Comment("optional program anchor for this metadata").
			Optional().
			Nillable(),
		field.String("platform_id").
			Comment("optional platform anchor for this metadata").
			Optional().
			Nillable(),
		field.String("system_name").
			Comment("system name used in OSCAL metadata").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("system_name"),
			),
		field.String("version").
			Comment("system version used in OSCAL metadata").
			Optional(),
		field.Text("description").
			Comment("system description used in OSCAL metadata").
			Optional(),
		field.Text("authorization_boundary").
			Comment("authorization boundary narrative for OSCAL export").
			Optional(),
		field.Enum("sensitivity_level").
			Comment("security sensitivity level of the system").
			GoType(enums.SystemSensitivityLevel("")).
			Default(enums.SystemSensitivityLevelUnknown.String()).
			Optional(),
		field.Time("last_reviewed").
			Comment("timestamp when metadata was last reviewed").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.JSON("revision_history", []any{}).
			Comment("structured revision history for OSCAL metadata").
			Optional().
			Annotations(
				entgql.Type("[Any!]"),
			),
		field.JSON("oscal_metadata_json", map[string]any{}).
			Comment("optional escape hatch for additional OSCAL metadata fields").
			Optional(),
	}
}

// Edges of the SystemMetadata
func (s SystemMetadata) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: s,
			edgeSchema: Program{},
			field:      "program_id",
			ref:        "system_metadata",
			comment:    "optional program this metadata belongs to",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: s,
			edgeSchema: Platform{},
			field:      "platform_id",
			ref:        "system_metadata",
			comment:    "optional platform this metadata belongs to",
		}),
	}
}

// Mixin of the SystemMetadata
func (s SystemMetadata) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "SMD",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[SystemMetadata](s,
				withParents(Program{}, Platform{}),
				withOrganizationOwner(true),
				//				withListObjectsFilter(),
			),
		},
	}.getMixins(s)
}

// Indexes of the SystemMetadata
func (SystemMetadata) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("program_id").
			Unique().
			Annotations(
				entsql.IndexWhere("deleted_at is NULL AND program_id is not NULL"),
			),
		index.Fields("platform_id").
			Unique().
			Annotations(
				entsql.IndexWhere("deleted_at is NULL AND platform_id is not NULL"),
			),
	}
}

// Modules returns modules required for this schema
func (SystemMetadata) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the SystemMetadata
//func (SystemMetadata) Annotations() []schema.Annotation {
//	return []schema.Annotation{
//		entfga.SelfAccessChecks(),
//		entx.NewExportable(
//			entx.WithOrgOwned(),
//		),
//	}
//}
//
//// Policy of the SystemMetadata
//func (s SystemMetadata) Policy() ent.Policy {
//	return policy.NewPolicy(
//		policy.WithMutationRules(
//			policy.CanCreateObjectsUnderParents([]string{
//				Program{}.PluralName(),
//				Platform{}.PluralName(),
//			}),
//			policy.CheckCreateAccess(),
//			policy.CheckOrgWriteAccess(),
//		),
//	)
//}
