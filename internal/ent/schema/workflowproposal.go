package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
)

// WorkflowProposal stores staged changes for a single object+domain that require approvals.
type WorkflowProposal struct {
	SchemaFuncs
	ent.Schema
}

const schemaWorkflowProposal = "workflow_proposal"

// Name returns the name of the WorkflowProposal schema.
func (WorkflowProposal) Name() string {
	return schemaWorkflowProposal
}

// GetType returns the type of the WorkflowProposal schema.
func (WorkflowProposal) GetType() any {
	return WorkflowProposal.Type
}

// PluralName returns the plural name of the WorkflowProposal schema.
func (WorkflowProposal) PluralName() string {
	return pluralize.NewClient().Plural(schemaWorkflowProposal)
}

// Fields of the WorkflowProposal.
func (WorkflowProposal) Fields() []ent.Field {
	return []ent.Field{
		field.String("workflow_object_ref_id").
			Comment("WorkflowObjectRef record that identifies the target object for this proposal").
			NotEmpty(),
		field.String("domain_key").
			Comment("Stable key representing the approval domain for this proposal").
			NotEmpty(),
		field.Enum("state").
			Comment("Current state of the proposal").
			GoType(enums.WorkflowProposalState("")).
			Default(string(enums.WorkflowProposalStateDraft)),
		field.Int("revision").
			Comment("Monotonic revision counter; incremented on edits").
			Default(1),
		field.JSON("changes", map[string]any{}).
			Comment("Staged field updates for this domain; applied only after approval").
			Optional(),
		field.String("proposed_hash").
			Comment("Hash of the current proposed changes for approval verification").
			Optional(),
		field.String("approved_hash").
			Comment("Hash of the proposed changes that satisfied approvals (what was approved)").
			Optional(),
		field.Time("submitted_at").
			Comment("Timestamp when this proposal was submitted for approval").
			Optional().
			Nillable(),
		field.String("submitted_by_user_id").
			Comment("User who submitted this proposal for approval").
			Optional(),
	}
}

// Edges of the WorkflowProposal.
func (w WorkflowProposal) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowObjectRef{},
			field:      "workflow_object_ref_id",
			comment:    "WorkflowObjectRef identifying the target object for this proposal",
			required:   true,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: w,
			edgeSchema: User{},
			name:       "submitted_by_user",
			field:      "submitted_by_user_id",
			comment:    "User who submitted this proposal for approval",
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: w,
			edgeSchema: WorkflowInstance{},
			name:       "workflow_instances",
			ref:        "workflow_proposal",
			comment:    "Workflow instances associated with this proposal",
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
	}
}

// Indexes of the WorkflowProposal.
func (WorkflowProposal) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("workflow_object_ref_id", "domain_key").
			Unique().
			Annotations(entsql.IndexWhere("((state)::text = ANY ((ARRAY['DRAFT'::character varying, 'SUBMITTED'::character varying])::text[]))")),
	}
}

// Mixin of the Integration
func (w WorkflowProposal) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeAnnotations: true,
		excludeSoftDelete:  true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.WorkflowObjectRef](w,
				withParents(WorkflowObjectRef{}),
				withOrganizationOwnerServiceOnly(true),
			),
		},
	}.getMixins(w)
}

// Modules this schema has access to.
func (WorkflowProposal) Modules() []models.OrgModule {
	return []models.OrgModule{models.CatalogBaseModule}
}

// Hooks returns the hooks for the WorkflowProposal schema
func (WorkflowProposal) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookWorkflowProposalInvalidateAssignments(),
		hooks.HookWorkflowProposalTriggerOnSubmit(),
	}
}

// Annotations returns the annotations for the WorkflowProposal schema
func (WorkflowProposal) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}
