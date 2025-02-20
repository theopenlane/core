package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	emixin "github.com/theopenlane/entx/mixin"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	// "github.com/theopenlane/iam/entfga"
)

// Evidence holds the schema definition for the Evidence entity
type Evidence struct {
	ent.Schema
}

// Fields of the Evidence
func (Evidence) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the evidence").
			NotEmpty(),
		field.String("description").
			Comment("the description of the evidence, what is contained in the uploaded file(s) or url(s)").
			Optional(),
		field.Text("collection_procedure").
			Comment("description of how the evidence was collected").
			Optional(),
		field.Time("creation_date").
			Comment("the date the evidence was retrieved").
			Default(time.Now),
		field.Time("renewal_date").
			Comment("the date the evidence should be renewed, defaults to a year from entry date").
			Default(time.Now().AddDate(1, 0, 0)).
			Optional(),
		field.String("source").
			Comment("the source of the evidence, e.g. system the evidence was retrieved from (splunk, github, etc)").
			Optional(),
		field.Bool("is_automated").
			Comment("whether the evidence was automatically generated").
			Optional().
			Default(false),
		field.String("url").
			Optional().
			Comment("the url of the evidence if not uploaded directly to the system"),
	}
}

// Mixin of the Evidence
func (Evidence) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.NewIDMixinWithPrefixedID("EVD"),
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
		NewObjectOwnedMixin(ObjectOwnedMixin{
			FieldNames:            []string{"control_id", "subcontrol_id", "control_objective_id", "program_id", "task_id"}, // used to create parent tuples for the evidence
			WithOrganizationOwner: true,
			Ref:                   "evidence",
		})}
}

// Edges of the Evidence
func (Evidence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("control_objectives", ControlObjective.Type),
		edge.To("controls", Control.Type),
		edge.To("subcontrols", Subcontrol.Type),
		edge.To("files", File.Type),
		edge.From("programs", Program.Type).
			Ref("evidence"),
		edge.From("tasks", Task.Type).
			Ref("evidence"),
	}
}

// Annotations of the Evidence
func (Evidence) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the Evidence
func (Evidence) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookEvidenceFiles(),
	}
}

// Policy of the Evidence
func (Evidence) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.EvidenceMutation](),
		),
	)
}
