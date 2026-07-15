package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
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
		defaultEdgeFromWithPagination(s, Program{}),
		defaultEdgeFromWithPagination(s, Platform{}),
		defaultEdgeFromWithPagination(s, Entity{}),
		defaultEdgeToWithPagination(s, Asset{}),
	}
}

// Mixin of the SystemDetail
func (s SystemDetail) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "SDT",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.SystemDetail](s,
				withParents(Program{}, Platform{}),
				withOrganizationOwner(),
				withSkipForSystemAdmin(),
			),
		},
	}.getMixins(s)
}

// Indexes of the SystemDetail
func (SystemDetail) Indexes() []ent.Index {
	return []ent.Index{}
}

// Modules returns modules required for this schema
func (SystemDetail) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogRegistryModule,
	}
}

// Annotations of the SystemDetail
func (SystemDetail) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.NewExportable(),
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
			policy.CheckOrgWriteAccess(),
			policy.CheckCreateAccess(),
			policy.CanCreateObjectsUnderParents([]string{
				Program{}.PluralName(),
				Platform{}.PluralName(),
			}),
		),
	)
}
