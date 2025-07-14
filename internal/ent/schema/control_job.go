package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
)

// ControlScheduledJob holds the schema definition for the ControlScheduledJob entity
type ControlScheduledJob struct {
	SchemaFuncs

	ent.Schema
}

// SchemaControlScheduledJobis the name of the schema in snake case
const SchemaControlScheduledJob = "scheduled_job"

func (ControlScheduledJob) Name() string {
	return SchemaControlScheduledJob
}

func (ControlScheduledJob) GetType() any {
	return ControlScheduledJob.Type
}

func (ControlScheduledJob) PluralName() string {
	return pluralize.NewClient().Plural(SchemaControlScheduledJob)
}

// Fields of the ControlScheduledJob
func (ControlScheduledJob) Fields() []ent.Field {
	return []ent.Field{
		field.String("job_id").
			Comment("the scheduled_job id to take the script to run from"),

		field.JSON("configuration", models.JobConfiguration{}).
			Annotations(
				entgql.Skip(entgql.SkipWhereInput | entgql.SkipOrderField),
			).
			Comment("the configuration to run this job"),

		field.JSON("cadence", models.JobCadence{}).
			Comment("the schedule to run this job. If not provided, it would inherit the cadence of the parent job").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput | entgql.SkipOrderField),
			).
			Optional(),

		field.String("cron").
			GoType(models.Cron("")).
			Comment("cron syntax. If not provided, it would inherit the cron of the parent job").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput |
					entgql.SkipOrderField,
				),
			).
			Optional().
			Nillable(),

		field.String("job_runner_id").
			Optional().
			Comment("the runner that this job will run on. If not set, it will scheduled on a general runner instead"),
	}
}

// Mixin of the ControlScheduledJob
func (c ControlScheduledJob) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(c),
		},
	}.getMixins()
}

// Edges of the ControlScheduledJob
func (c ControlScheduledJob) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: c,
			edgeSchema: ScheduledJob{},
			field:      "job_id",
			required:   true,
		}),

		defaultEdgeToWithPagination(c, Control{}),
		defaultEdgeToWithPagination(c, Subcontrol{}),

		uniqueEdgeTo(&edgeDefinition{
			fromSchema: c,
			edgeSchema: JobRunner{},
			field:      "job_runner_id",
			required:   false,
		}),
	}
}

// Indexes of the ControlScheduledJob
func (ControlScheduledJob) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ControlScheduledJob
func (ControlScheduledJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features("compliance", "continuous-compliance-automation"),
	}
}

// Hooks of the ControlScheduledJob
func (ControlScheduledJob) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookControlScheduledJobCreate(),
	}
}

// Interceptors of the ControlScheduledJob
func (ControlScheduledJob) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the ControlScheduledJob
func (ControlScheduledJob) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
