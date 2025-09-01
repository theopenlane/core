package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// InternalPolicy defines the policy schema.
type InternalPolicy struct {
	SchemaFuncs

	ent.Schema
}

// SchemaInternalPolicy is the name of the internal policy schema.
const SchemaInternalPolicy = "internal_policy"

// Name returns the name of the internal policy schema.
func (InternalPolicy) Name() string {
	return SchemaInternalPolicy
}

// GetType returns the type of the internal policy schema.
func (InternalPolicy) GetType() any {
	return InternalPolicy.Type
}

// PluralName returns the plural name of the internal policy schema.
func (InternalPolicy) PluralName() string {
	return pluralize.NewClient().Plural(SchemaInternalPolicy)
}

// Fields returns policy fields.
func (InternalPolicy) Fields() []ent.Field {
	// other fields are defined in the mixins
	return []ent.Field{
		field.String("file_id").
			Comment("This will contain the most recent file id if this policy was created from a file").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Nillable(),

		field.String("url").
			Comment("This will contain the url used to create/update the policy").
			Optional().
			Nillable(),
	}
}

// Edges of the InternalPolicy
func (i InternalPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(i, ControlObjective{}),
		defaultEdgeToWithPagination(i, ControlImplementation{}),
		defaultEdgeToWithPagination(i, Control{}),
		defaultEdgeToWithPagination(i, Subcontrol{}),
		defaultEdgeToWithPagination(i, Procedure{}),
		defaultEdgeToWithPagination(i, Narrative{}),
		defaultEdgeToWithPagination(i, Task{}),
		defaultEdgeToWithPagination(i, Risk{}),

		defaultEdgeFromWithPagination(i, Program{}),

		uniqueEdgeTo(&edgeDefinition{
			fromSchema: i,
			edgeSchema: File{},
			field:      "file_id",
		}),
	}
}

// Mixin of the InternalPolicy
func (i InternalPolicy) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:          "PLC",
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			// all policies must be associated to an organization
			newOrgOwnedMixin(i,
				withSkipForSystemAdmin(true),
			),
			// add group edit permissions to the procedure
			newGroupPermissionsMixin(withSkipViewPermissions(), withGroupPermissionsInterceptor()),
			// policies are documents
			DocumentMixin{DocumentType: "policy"}, // use short name for the document type
		},
	}.getMixins(i)
}

func (InternalPolicy) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogPolicyManagementAddon,
	}
}

// Annotations of the InternalPolicy
func (i InternalPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.Exportable{},
	}
}

// Hooks of the InternalPolicy
func (InternalPolicy) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookPolicy(),
		hook.On(
			hooks.OrgOwnedTuplesHookWithAdmin(),
			ent.OpCreate,
		),
	}
}

// Interceptors of the InternalPolicy
func (InternalPolicy) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		// policies are org owned, but we need to ensure the groups are filtered as well
		interceptors.FilterQueryResults[generated.InternalPolicy](),
	}
}

// Policy of the InternalPolicy
func (i InternalPolicy) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{
				Program{}.PluralName(),
			}),
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.InternalPolicyMutation](),
		),
	)
}
