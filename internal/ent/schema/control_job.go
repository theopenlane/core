package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx/history"
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
		field.String("job_id"),

		field.JSON("configuration", models.JobConfiguration{}).
			Annotations(
				entgql.Skip(entgql.SkipWhereInput | entgql.SkipOrderField),
			).
			Comment("the configuration to run this job"),

		field.JSON("cadence", models.JobCadence{}).
			Comment("the schedule to run this job").
			Annotations(
				entgql.Skip(entgql.SkipWhereInput | entgql.SkipOrderField),
			).
			Optional(),

		field.String("cron").
			Comment("cron syntax").
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
		// entfga.SelfAccessChecks(),
		entgql.RelayConnection(),
		history.Annotations{
			Exclude: true,
		},
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
	// add the new policy here, the default post-policy is to deny all
	// so you need to ensure there are rules in place to allow the actions you want
	return policy.NewPolicy(
		policy.WithQueryRules(
			// add query rules here, the below is the recommended default
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			// add mutation rules here, the below is the recommended default
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
