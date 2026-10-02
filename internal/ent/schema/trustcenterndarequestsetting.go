package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/interceptors"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
)

// TrustCenterNDARequestSetting holds the schema definition for the TrustCenterNDARequestSetting entity
type TrustCenterNDARequestSetting struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustCenterNDARequestSetting is the name of the schema in snake case
const SchemaTrustCenterNDARequestSetting = "trust_center_nda_request_setting"

// Name is the name of the schema in snake case
func (TrustCenterNDARequestSetting) Name() string {
	return SchemaTrustCenterNDARequestSetting
}

// GetType returns the type of the schema
func (TrustCenterNDARequestSetting) GetType() any {
	return TrustCenterNDARequestSetting.Type
}

// PluralName returns the plural name of the schema
func (TrustCenterNDARequestSetting) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenterNDARequestSetting)
}

// Fields of the TrustCenterNDARequestSetting
func (TrustCenterNDARequestSetting) Fields() []ent.Field {
	return []ent.Field{
		field.String("approver_group_id").
			Comment("group whose members approve trust center NDA requests").
			Optional().
			Nillable(),
		field.Bool("approval_required").
			Comment("whether NDA requests require approval before being processed").
			Default(false).
			Optional(),
		field.Bool("auto_approve").
			Comment("Auto approve NDA requests based on certain rules").
			Default(false).
			Optional(),
		field.Bool("work_email_only").
			Comment("require only work email when auto approving a nda request").
			Default(true).
			Optional(),
		field.Bool("use_domain_blocklist").
			Comment("Enable the use of a domain blocklist").
			Default(false).
			Optional(),
		field.Bool("use_domain_allowlist").
			Comment("Enable the use of a domain allowlist").
			Default(false).
			Optional(),
		field.Bool("approve_from_existing_domain").
			Comment("automatically approve if the request uses the same domain as that of an existing domain").
			Default(true).
			Optional(),
		field.Bool("approve_if_contact_exists").
			Comment("automatically approve if the request email is already a contact object").
			Default(true).
			Optional(),
	}
}

// Mixin of the TrustCenterNDARequestSetting
func (t TrustCenterNDARequestSetting) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.TrustCenterSetting](t,
				withParents(TrustCenter{}),
			),
			newGroupPermissionsMixin(withSkipViewPermissions()),
		},
	}.getMixins(t)
}

// Edges of the TrustCenterNDARequestSetting
func (t TrustCenterNDARequestSetting) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			name:       "approver_group",
			t:          Group.Type,
			field:      "approver_group_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Group{}.Name()),
			},
		}),
	}
}

// Indexes of the TrustCenterNDARequestSetting
func (TrustCenterNDARequestSetting) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the TrustCenterNDARequestSetting
func (TrustCenterNDARequestSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// entfga.SettingsChecks("trust_center"),
		entgql.QueryField("trustCenterNDARequestSettings"),
		entx.FileCategory(SchemaTrustCenterSetting),
		entx.FGACrudSkip(entx.SkipDelete | entx.SkipCreate),
	}
}

// Hooks of the TrustCenterNDARequestSetting
func (TrustCenterNDARequestSetting) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the TrustCenterNDARequestSetting
func (TrustCenterNDARequestSetting) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}

// Modules this schema has access to
func (TrustCenterNDARequestSetting) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Policy of the TrustCenterNDARequestSetting
func (TrustCenterNDARequestSetting) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithOnMutationRules(ent.OpCreate,
			rule.AllowIfTrustCenterEditor(),
			policy.CanCreateObjectsUnderParents([]string{
				TrustCenter{}.Name(),
			}),
			policy.CheckOrgWriteAccess(),
		),
		policy.WithOnMutationRules(ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne,
			rule.AllowIfTrustCenterEditor(),
			// entfga.CheckEditAccess[*generated.TrustCenterSettingMutation](),
		),
	)
}
