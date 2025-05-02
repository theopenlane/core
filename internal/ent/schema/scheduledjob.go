package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"
)

// ScheduledJob holds the schema definition for the ScheduledJob entity
type ScheduledJob struct {
	SchemaFuncs

	ent.Schema
}

// SchemaScheduledJobis the name of the schema in snake case
const SchemaScheduledJob = "scheduled_job"

func (ScheduledJob) Name() string {
	return SchemaScheduledJob
}

func (ScheduledJob) GetType() any {
	return ScheduledJob.Type
}

func (ScheduledJob) PluralName() string {
	return pluralize.NewClient().Plural(SchemaScheduledJob)
}

// Fields of the ScheduledJob
func (ScheduledJob) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").
			Comment("the title of the task").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("title"),
			).
			NotEmpty(),
		field.String("description").
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("the description of the task").
			Optional(),
		field.Enum("job_type").
			GoType(enums.JobType("")).
			Default(enums.JobTypeSsl.String()).
			Annotations(
				entgql.OrderField("job_type"),
				entgql.Skip(
					entgql.SkipMutationCreateInput|
						entgql.SkipMutationUpdateInput,
				),
			).
			Comment("the type of this job"),

		field.Enum("environment").
			GoType(enums.JobEnvironment("")).
			Default(enums.JobEnvironmentOpenlane.String()).
			Annotations(
				entgql.OrderField("environment"),
				entgql.Skip(
					entgql.SkipMutationCreateInput|
						entgql.SkipMutationUpdateInput,
				),
			).
			Comment("the type of this job"),

		field.String("script").
			Annotations(
				entgql.Skip(^entgql.SkipType),
			).
			Comment("the script to run").
			Optional(),

		field.Bool("is_active").
			Annotations(
				entgql.Skip(
					entgql.SkipMutationCreateInput |
						entgql.SkipMutationUpdateInput,
				),
			).
			Default(true),
	}
}

// Mixin of the ScheduledJob
func (s ScheduledJob) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "JOB",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s,
				withSkipForSystemAdmin(true),
			),
			mixin.SystemOwnedMixin{},
		},
	}.getMixins()
}

// Edges of the ScheduledJob
func (s ScheduledJob) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema:    s,
			name:          "scheduled_job_setting",
			t:             ScheduledJobSetting.Type,
			cascadeDelete: "ScheduledJob",
			required:      true,
		}),
		// defaultEdgeFromWithPagination(s, Control{}),
	}
}

// Indexes of the ScheduledJob
func (ScheduledJob) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ScheduledJob
func (ScheduledJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Hooks of the ScheduledJob
func (ScheduledJob) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Policy of the ScheduledJob
func (ScheduledJob) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAndDeleteAccess[*generated.ScheduledJobMutation](),
		),
	)
}
