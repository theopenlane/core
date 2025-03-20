package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"
)

// Program holds the schema definition for the Program entity
type Program struct {
	SchemaFuncs

	ent.Schema
}

const SchemaProgram = "program"

func (Program) Name() string {
	return SchemaProgram
}

func (Program) GetType() any {
	return Program.Type
}

func (Program) PluralName() string {
	return pluralize.NewClient().Plural(SchemaProgram)
}

// Fields of the Program
func (Program) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the program").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("the description of the program").
			Annotations(
				entx.FieldSearchable(),
			).
			Optional(),
		field.Enum("status").
			Comment("the status of the program").
			GoType(enums.ProgramStatus("")).
			Annotations(
				entgql.OrderField("STATUS"),
			).
			Default(enums.ProgramStatusNotStarted.String()),
		field.Time("start_date").
			Comment("the start date of the period").
			Annotations(
				entgql.OrderField("start_date"),
			).
			Optional(),
		field.Time("end_date").
			Comment("the end date of the period").
			Annotations(
				entgql.OrderField("end_date"),
			).
			Optional(),
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
	return mixinConfig{
		prefix: "PRG",
		additionalMixins: []ent.Mixin{
			// all programs must be associated to an organization
			NewOrgOwnMixinWithRef("programs"),
			// add group permissions to the program
			NewGroupPermissionsMixin(true),
		},
	}.getMixins()
}

// Edges of the Program
func (p Program) Edges() []ent.Edge {
	return []ent.Edge{
		// programs can have 1:many controls
		defaultEdgeToWithPagination(p, Control{}),
		// programs can have 1:many subcontrols
		defaultEdgeToWithPagination(p, Subcontrol{}),
		// programs can have 1:many control objectives
		defaultEdgeToWithPagination(p, ControlObjective{}),
		// programs can have 1:many associated policies
		defaultEdgeToWithPagination(p, InternalPolicy{}),
		// programs can have 1:many associated procedures
		defaultEdgeToWithPagination(p, Procedure{}),
		// programs can have 1:many associated risks
		defaultEdgeToWithPagination(p, Risk{}),
		// programs can have 1:many associated tasks
		defaultEdgeToWithPagination(p, Task{}),
		// programs can have 1:many associated notes (comments)
		defaultEdgeToWithPagination(p, Note{}),
		// programs can have 1:many associated files
		defaultEdgeToWithPagination(p, File{}),
		// programs can be many:many with evidence
		defaultEdgeToWithPagination(p, Evidence{}),
		// programs can have 1:many associated narratives
		defaultEdgeToWithPagination(p, Narrative{}),
		// programs can have 1:many associated action plans
		defaultEdgeToWithPagination(p, ActionPlan{}),
		edge.From("users", User.Type).
			Ref("programs").
			// Skip the mutation input for the users edge
			// this should be done via the members edge
			Annotations(entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput), entgql.RelayConnection()).
			Through("members", ProgramMembership.Type),
	}
}

// Annotations of the Program
func (Program) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// Delete groups members when groups are deleted
		entx.CascadeThroughAnnotationField(
			[]entx.ThroughCleanup{
				{
					Field:   "Program",
					Through: "ProgramMembership",
				},
			},
		),
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the Program
func (Program) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookProgramAuthz(),
	}
}

// Interceptors of the Program
func (Program) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.FilterListQuery(),
	}
}

// Policy of the program
func (Program) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ProgramMutation](),
		),
	)
}
