package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/shared/models"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/entx/accessmap"
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
	}.getMixins(s)
}

// Edges of the ScheduledJobRun
func (s ScheduledJobRun) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: s,
			edgeSchema: ScheduledJob{},
			field:      "scheduled_job_id",
			required:   true,
		}),

		uniqueEdgeTo(&edgeDefinition{
			fromSchema: s,
			edgeSchema: JobRunner{},
			field:      "job_runner_id",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Organization{}.Name()),
			},
		}),
	}
}

// Indexes of the ScheduledJobRun
func (ScheduledJobRun) Indexes() []ent.Index {
	return []ent.Index{}
}

func (ScheduledJobRun) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the ScheduledJobRun
func (s ScheduledJobRun) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the ScheduledJobRun
func (ScheduledJobRun) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Policy of the ScheduledJobRun
func (s ScheduledJobRun) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
