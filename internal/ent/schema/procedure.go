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

// Procedure defines the procedure schema.
type Procedure struct {
	SchemaFuncs

	ent.Schema
}

// SchemaProcedure is the name of the procedure schema.
const SchemaProcedure = "procedure"

// Name returns the name of the procedure schema.
func (Procedure) Name() string {
	return SchemaProcedure
}

// GetType returns the type of the procedure schema.
func (Procedure) GetType() any {
	return Procedure.Type
}

// PluralName returns the plural name of the procedure schema.
func (Procedure) PluralName() string {
	return pluralize.NewClient().Plural(SchemaProcedure)
}

// Fields returns procedure fields.
func (Procedure) Fields() []ent.Field {
	// other fields are defined in the mixins
	return []ent.Field{
		field.String("file_id").
			Comment("This will contain the most recent file id if this procedure was created from a file").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Nillable(),

		field.String("url").
			Comment("This will contain the url used to create/update the procedure").
			Optional().
			Nillable(),
	}
}

// Edges of the Procedure
func (p Procedure) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(p, Control{}),
		defaultEdgeFromWithPagination(p, Subcontrol{}),
		defaultEdgeFromWithPagination(p, InternalPolicy{}),
		defaultEdgeFromWithPagination(p, Program{}),
		defaultEdgeToWithPagination(p, Narrative{}),
		defaultEdgeToWithPagination(p, Risk{}),
		defaultEdgeToWithPagination(p, Task{}),

		uniqueEdgeTo(&edgeDefinition{
			fromSchema: p,
			edgeSchema: File{},
			field:      "file_id",
		}),
	}
}

// Mixin of the Procedure
func (p Procedure) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:          "PRD",
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(p, withSkipForSystemAdmin(true)),
			// add group edit permissions to the procedure
			newGroupPermissionsMixin(withSkipViewPermissions(), withGroupPermissionsInterceptor()),
			// all procedures are documents
			NewDocumentMixin(p),
		},
	}.getMixins(p)
}

func (Procedure) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogPolicyManagementAddon,
		models.CatalogRiskManagementAddon,
		models.CatalogEntityManagementModule,
	}
}

// Annotations of the Procedure
func (p Procedure) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.Exportable{},
	}
}

// Hooks of the Procedure
func (Procedure) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookProcedure(),
		hook.On(
			hooks.OrgOwnedTuplesHookWithAdmin(),
			ent.OpCreate,
		),
	}
}

// Interceptors of the Procedure
func (p Procedure) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		// procedures are org owned, but we need to ensure the groups are filtered as well
		interceptors.FilterQueryResults[generated.Procedure](),
	}
}

// Policy of the Procedure
func (p Procedure) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(),
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{
				Program{}.PluralName(),
			}),
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ProcedureMutation](),
		),
	)
}
