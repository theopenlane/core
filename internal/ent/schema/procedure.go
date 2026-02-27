package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
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
	return []ent.Field{}
}

// Edges of the Procedure
func (p Procedure) Edges() []ent.Edge {
	return []ent.Edge{
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: p,
			edgeSchema: Control{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: p,
			edgeSchema: Subcontrol{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: p,
			edgeSchema: InternalPolicy{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: p,
			edgeSchema: Program{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: p,
			edgeSchema: Narrative{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: p,
			edgeSchema: Risk{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: p,
			edgeSchema: Task{},
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: p,
			name:       "comments",
			t:          Note.Type,
			comment:    "conversations related to the procedure",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Note{}.Name()),
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: p,
			edgeSchema: Discussion{},
			comment:    "discussions related to the procedure",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Note{}.Name()),
				entx.FieldWorkflowEligible(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: p,
			edgeSchema: File{},
			field:      "file_id",
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: p,
			edgeSchema: WorkflowObjectRef{},
			name:       "workflow_object_refs",
			ref:        "procedure",
			annotations: []schema.Annotation{
				entx.FieldWorkflowEligible(),
			},
		}),
	}
}

// Mixin of the Procedure
func (p Procedure) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:          "PRD",
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(p),
			// add group edit permissions to the procedure
			newGroupPermissionsMixin(withSkipViewPermissions(), withGroupPermissionsInterceptor(), withWorkflowGroupEdges()),
			// all procedures are documents
			NewDocumentMixin(p),
			mixin.NewSystemOwnedMixin(mixin.SkipTupleCreation()),
			newCustomEnumMixin(p, withWorkflowEnumEdges()),
			newCustomEnumMixin(p, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(p, withEnumFieldName("scope"), withGlobalEnum()),
			WorkflowApprovalMixin{},
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
		entx.NewExportable(
			entx.WithOrgOwned(),
		),
	}
}

// Hooks of the Procedure
func (Procedure) Hooks() []ent.Hook {
	return []ent.Hook{
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
			rule.CheckIfCommentOnly(),
			entfga.CheckEditAccess[*generated.ProcedureMutation](),
		),
	)
}
