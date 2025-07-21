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

// ScheduledJob holds the schema definition for the ScheduledJob entity
type ScheduledJob struct {
	SchemaFuncs

	ent.Schema
}

// SchemaScheduledJob is the name of the schema in snake case
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

// Fields of the ControlScheduledJob
func (ScheduledJob) Fields() []ent.Field {
	return []ent.Field{
		field.String("job_id").
			Comment("the scheduled_job id to take the script to run from"),

		field.JSON("configuration", models.JobConfiguration{}).
			Annotations(
				entgql.Skip(entgql.SkipWhereInput | entgql.SkipOrderField),
			).
			Optional().
			Comment("the configuration to run this job"),
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
func (c ScheduledJob) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:      "SJB",
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			// TODO: update to object owned mixin
			// that will allow inheritance by control + subcontrol to edit the scheduled job
			// currently only org admins can create and edit scheduled jobs
			newOrgOwnedMixin(c),
		},
	}.getMixins()
}

// Edges of the ControlScheduledJob
func (c ScheduledJob) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: c,
			edgeSchema: JobTemplate{},
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
func (ScheduledJob) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ControlScheduledJob
func (ScheduledJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features("compliance", "continuous-compliance-automation"),
		entx.SchemaSearchable(false), // not sure yet why, but this breaks generation when enabled
	}
}

// Hooks of the ControlScheduledJob
func (ScheduledJob) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookScheduledJobCreate(),
	}
}

// Interceptors of the ControlScheduledJob
func (ScheduledJob) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the ControlScheduledJob
func (ScheduledJob) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
