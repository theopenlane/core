package schema

import (
	"context"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Subcontrol defines the file schema.
type Subcontrol struct {
	ent.Schema
}

// Fields returns file fields.
func (Subcontrol) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("the name of the subcontrol"),
		field.Text("description").
			Optional().
			Comment("description of the subcontrol"),
		field.String("status").
			Optional().
			Comment("status of the subcontrol"),
		field.String("subcontrol_type").
			Optional().
			Comment("type of the subcontrol"),
		field.String("version").
			Optional().
			Comment("version of the control"),
		field.String("subcontrol_number").
			Optional().
			Comment("number of the subcontrol"),
		field.Text("family").
			Optional().
			Comment("subcontrol family"),
		field.String("class").
			Optional().
			Comment("subcontrol class"),
		field.String("source").
			Optional().
			Comment("source of the control, e.g. framework, template, user-defined, etc."),
		field.Text("mapped_frameworks").
			Optional().
			Comment("mapped frameworks that the subcontrol is part of"),
		field.String("implementation_evidence").
			Optional().
			Comment("implementation evidence of the subcontrol"),
		field.String("implementation_status").
			Optional().
			Comment("implementation status"),
		field.Time("implementation_date").
			Optional().
			Comment("date the subcontrol was implemented"),
		field.String("implementation_verification").
			Optional().
			Comment("implementation verification"),
		field.Time("implementation_verification_date").
			Optional().
			Comment("date the subcontrol implementation was verified"),
		field.JSON("details", map[string]interface{}{}).
			Optional().
			Comment("json data details of the subcontrol"),
	}
}

// Edges of the Subcontrol
func (Subcontrol) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("control", Control.Type).
			Ref("subcontrols"),
		edge.To("tasks", Task.Type),
		edge.From("notes", Note.Type).
			Unique().
			Ref("subcontrols"),
		edge.From("programs", Program.Type).
			Ref("subcontrols"),
	}
}

// Mixin of the Subcontrol
func (Subcontrol) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		mixin.SoftDeleteMixin{},
		emixin.IDMixin{},
		emixin.TagMixin{},
		// subcontrols can inherit permissions from the program
		// and must be owned by an organization
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:            []string{"program_id"},
			WithOrganizationOwner: true,
			Ref:                   "subcontrols",
		}),
		// add groups permissions with viewer, editor, and blocked groups
		NewGroupPermissionsMixin(true),
	}
}

// Annotations of the Subcontrol
func (Subcontrol) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.Annotations{
			ObjectType: "subcontrol", // check access to the risk for update/delete
		},
	}
}

// Policy of the Subcontrol
func (Subcontrol) Policy() ent.Policy {
	return privacy.Policy{
		Mutation: privacy.MutationPolicy{
			rule.CanCreateObjectsInProgram(), // if mutation contains program_id, check access
			privacy.OnMutationOperation( // if there is no program_id, check access for create in org
				rule.CanCreateObjectsInOrg(),
				ent.OpCreate,
			),
			privacy.SubcontrolMutationRuleFunc(func(ctx context.Context, m *generated.SubcontrolMutation) error {
				return m.CheckAccessForEdit(ctx) // check access for edit
			}),
			privacy.AlwaysDenyRule(),
		},
		Query: privacy.QueryPolicy{
			privacy.SubcontrolQueryRuleFunc(func(ctx context.Context, q *generated.SubcontrolQuery) error {
				return q.CheckAccess(ctx)
			}),
			privacy.AlwaysDenyRule(),
		},
	}
}
