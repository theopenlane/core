package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
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
		// subcontorls have to have a parent control
		edge.From("controls", Control.Type).
			Required().
			Ref("subcontrols"),
		edge.From("user", User.Type).
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
		// subcontrols can inherit permissions from the parent control
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:            []string{"control_id"},
			WithOrganizationOwner: true,
			Ref:                   "subcontrols",
		}),
	}
}

// Annotations of the Subcontrol
func (Subcontrol) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.RelayConnection(),
		entgql.QueryField(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the Subcontrol
func (Subcontrol) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookSubcontrolUpdate(),
	}
}

// Policy of the Subcontrol
func (Subcontrol) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			entfga.CheckReadAccess[*generated.SubcontrolQuery](),
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.SubcontrolMutation](rule.ControlParent), // if mutation contains control_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.SubcontrolMutation](),
		),
	)
}
