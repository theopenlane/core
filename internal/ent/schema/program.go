package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/pkg/enums"
	emixin "github.com/theopenlane/entx/mixin"
)

// Program holds the schema definition for the Program entity
type Program struct {
	ent.Schema
}

// Fields of the Program
func (Program) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the program").
			NotEmpty(),
		field.String("description").
			Comment("the description of the program").
			Optional(),
		field.Enum("status").
			Comment("the status of the program").
			GoType(enums.ProgramStatus("")).
			Default(enums.ProgramStatusNotStarted.String()),
		field.Time("start_date").
			Comment("the start date of the period").
			Optional(),
		field.Time("end_date").
			Comment("the end date of the period").
			Optional(),
		field.String("organization_id").
			Comment("the organization that owns the program").
			NotEmpty(),
		field.Bool("auditor_ready").
			Comment("is the program ready for the auditor").
			Default(false),
		field.Bool("auditor_write_comments").
			Comment("can the auditor write comments").
			Default(false),
		field.Bool("auditor_read_comments").
			Comment("can the auditor read comments").
			Default(false),
	}
}

// Mixin of the Program
func (Program) Mixin() []ent.Mixin {
	return []ent.Mixin{
		emixin.AuditMixin{},
		emixin.IDMixin{},
		mixin.SoftDeleteMixin{},
		emixin.TagMixin{},
	}
}

// Edges of the Program
func (Program) Edges() []ent.Edge {
	return []ent.Edge{
		// all programs must be associated to an organization
		edge.From("organization", Organization.Type).
			Ref("programs").
			Field("organization_id").
			Required().
			Unique(),
		// programs can have 1:many controls
		edge.To("controls", Control.Type),
		// programs can have 1:many subcontrols
		edge.To("subcontrols", Subcontrol.Type),
		// programs can have 1:many control objectives
		edge.To("controlobjectives", ControlObjective.Type),
		// programs can have 1:many associated policies
		edge.To("policies", InternalPolicy.Type),
		// programs can have 1:many associated procedures
		edge.To("procedures", Procedure.Type),
		// programs can have 1:many associated risks
		edge.To("risks", Risk.Type),
		// programs can have 1:many associated tasks
		edge.To("tasks", Task.Type),
		// programs can have 1:many associated notes (comments)
		edge.To("notes", Note.Type),
		// programs can have 1:many associated files
		edge.To("files", File.Type),
		// programs can have 1:many associated narratives
		edge.To("narratives", Narrative.Type),
		// programs can have 1:many associated action plans
		edge.To("actionplans", ActionPlan.Type),
		// programs can have 1:many associated standards (frameworks)
		edge.From("standards", Standard.Type).
			Ref("programs").
			Comment("the framework(s) that the program is based on"),
		edge.From("users", User.Type).
			Ref("programs").
			Through("members", ProgramMembership.Type),
	}
}

// Indexes of the Program
func (Program) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the Program
func (Program) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entgql.RelayConnection(),
		entgql.Mutations(entgql.MutationCreate(), (entgql.MutationUpdate())),
	}
}

// Hooks of the Program
func (Program) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the Program
func (Program) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}
