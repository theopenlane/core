package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
)

// ScheduledJobRun holds the schema definition for the ScheduledJobRun entity
type ScheduledJobRun struct {
	SchemaFuncs

	ent.Schema
}

// SchemaScheduledJobRunis the name of the schema in snake case
const SchemaScheduledJobRun = "scheduled_job_run"

// Name is the name of the schema in snake case
func (ScheduledJobRun) Name() string {
	return SchemaScheduledJobRun
}

// GetType returns the type of the schema
func (ScheduledJobRun) GetType() any {
	return ScheduledJobRun.Type
}

// PluralName returns the plural name of the schema
func (ScheduledJobRun) PluralName() string {
	return pluralize.NewClient().Plural(SchemaScheduledJobRun)
}

// Fields of the ScheduledJobRun
func (ScheduledJobRun) Fields() []ent.Field {
	return []ent.Field{
		field.String("job_runner_id").
			Comment("The runner that this job will be executed on. Useful to know because of self hosted runners"),

		field.Enum("status").
			GoType(enums.ScheduledJobRunStatus("")).
			Default(enums.ScheduledJobRunStatusPending.String()).
			Comment(`The status of the job to be executed. By default will be pending but when
			scheduled on a runner, this will change to acquired.`),

		field.String("scheduled_job_id").
			Comment("the parent job for this run"),

		field.Time("expected_execution_time").
			Immutable().
			Comment("When should this job execute on the agent. Since we might potentially schedule a few minutes before"),

		field.String("script").
			Immutable().
			// The script in the job allows for templating so you
			// can do something like {{ .URL }}
			// Then when the job is being scheduled, it would replace with actual values
			// using text/template .
			//
			// This complete value is what can be executed with all required inputs
			Comment(`the script that will be executed by the agent.
This script will be templated with the values from the configuration on the job`),
	}
}

// Mixin of the ScheduledJobRun
func (s ScheduledJobRun) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s),
		},
	}.getMixins()
}

// Edges of the ScheduledJobRun
func (s ScheduledJobRun) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: s,
			edgeSchema: ControlScheduledJob{},
			field:      "scheduled_job_id",
			required:   true,
		}),

		uniqueEdgeTo(&edgeDefinition{
			fromSchema: s,
			edgeSchema: JobRunner{},
			field:      "job_runner_id",
			required:   true,
		}),
	}
}

// Indexes of the ScheduledJobRun
func (ScheduledJobRun) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ScheduledJobRun
func (ScheduledJobRun) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features("compliance", "continuous-compliance-automation"),
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the ScheduledJobRun
func (ScheduledJobRun) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the ScheduledJobRun
func (ScheduledJobRun) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the ScheduledJobRun
func (ScheduledJobRun) Policy() ent.Policy {
	// add the new policy here, the default post-policy is to deny all
	// so you need to ensure there are rules in place to allow the actions you want
	return policy.NewPolicy(
		policy.WithQueryRules(
			// add query rules here, the below is the recommended default
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
		),
	)
}
