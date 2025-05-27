package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
)

// ScheduledJob holds the schema definition for the ScheduledJob entity
type ScheduledJob struct {
	SchemaFuncs

	ent.Schema
}

// SchemaScheduledJobis the name of the schema in snake case
const SchemaScheduledJob = "job"

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
			Comment("the title of the job").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("title"),
			).
			NotEmpty(),
		field.String("description").
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("the description of the job").
			Optional(),
		field.Enum("job_type").
			GoType(enums.JobType("")).
			Default(enums.JobTypeSsl.String()).
			Annotations(
				entgql.OrderField("JOB_TYPE"),
			).
			Comment("the type of this job"),

		field.String("script").
			Annotations(
				entgql.Skip(
					entgql.SkipOrderField |
						entgql.SkipWhereInput,
				),
			).
			Comment("the script to run").
			Optional(),

		// Default config values
		field.JSON("configuration", models.JobConfiguration{}).
			Comment("the configuration to run this job"),

		field.JSON("cadence", models.JobCadence{}).
			Comment("the schedule to run this job").
			Optional(),

		field.String("cron").
			GoType(models.Cron("")).
			Comment("cron syntax").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput | entgql.SkipOrderField),
			).
			Optional().
			Nillable(),
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
	return []ent.Edge{}
}

// Indexes of the ScheduledJob
func (ScheduledJob) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ScheduledJob
func (ScheduledJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// entfga.SelfAccessChecks(),
	}
}

// Hooks of the ScheduledJob
func (ScheduledJob) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookScheduledJobCreate(),
	}
}

// Interceptors of the ScheduledJob
func (ScheduledJob) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the ScheduledJob
func (ScheduledJob) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
