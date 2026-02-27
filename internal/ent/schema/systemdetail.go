package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/oscalgen"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// SystemDetail defines OSCAL-centric system metadata anchors
type SystemDetail struct {
	SchemaFuncs

	ent.Schema
}

// SchemaSystemDetail is the canonical schema name
const SchemaSystemDetail = "system_detail"

// Name returns the schema name
func (SystemDetail) Name() string {
	return SchemaSystemDetail
}

// GetType returns the ent type
func (SystemDetail) GetType() any {
	return SystemDetail.Type
}

// PluralName returns the plural schema name
func (SystemDetail) PluralName() string {
	return "system_details"
}

// Fields of the SystemDetail
func (SystemDetail) Fields() []ent.Field {
	return []ent.Field{
		field.String("program_id").
			Comment("optional program anchor for this system detail").
			Optional().
			Nillable(),
		field.String("platform_id").
			Comment("optional platform anchor for this system detail").
			Optional().
			Nillable(),
		field.String("system_name").
			Comment("system name used in OSCAL metadata").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("system_name"),
				oscalgen.NewOSCALField(
					oscalgen.OSCALFieldRoleSystemName,
					oscalgen.WithOSCALFieldModels(oscalgen.OSCALModelSSP),
				),
			),
		field.String("version").
			Comment("system version used in OSCAL metadata").
			Optional(),
		field.Text("description").
			Comment("system description used in OSCAL metadata").
			Optional().
			Annotations(
				oscalgen.NewOSCALField(
					oscalgen.OSCALFieldRoleDescription,
					oscalgen.WithOSCALFieldModels(oscalgen.OSCALModelSSP),
				),
			),
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

// Edges of the SystemDetail
func (s SystemDetail) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: s,
			edgeSchema: Program{},
			field:      "program_id",
			ref:        "system_detail",
			comment:    "optional program this detail belongs to",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: s,
			edgeSchema: Platform{},
			field:      "platform_id",
			ref:        "system_detail",
			comment:    "optional platform this detail belongs to",
		}),
	}
}

// Mixin of the SystemDetail
func (s SystemDetail) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "SDT",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.SystemDetail](s,
				withParents(Program{}, Platform{}),
				withOrganizationOwner(true),
			),
		},
	}.getMixins(s)
}

// Indexes of the SystemDetail
func (SystemDetail) Indexes() []ent.Index {
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
func (SystemDetail) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the SystemDetail
func (SystemDetail) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.NewExportable(
			entx.WithOrgOwned(),
		),
		oscalgen.NewOSCALModel(
			oscalgen.WithOSCALModels(oscalgen.OSCALModelSSP),
			oscalgen.WithOSCALAssembly("system-characteristics"),
		),
	}
}

// Policy of the SystemDetail
func (s SystemDetail) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{
				Program{}.PluralName(),
				Platform{}.PluralName(),
			}),
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
