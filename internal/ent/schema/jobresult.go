package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/entx/history"
)

// JobResult holds the schema definition for the JobResult entity
type JobResult struct {
	SchemaFuncs

	ent.Schema
}

// SchemaJobResultis the name of the schema in snake case
const SchemaJobResult = "scheduled_job_result"

// Name is the name of the schema in snake case
func (JobResult) Name() string {
	return SchemaJobResult
}

// GetType returns the type of the schema
func (JobResult) GetType() any {
	return JobResult.Type
}

// PluralName returns the plural name of the schema
func (JobResult) PluralName() string {
	return pluralize.NewClient().Plural(SchemaJobResult)
}

// Fields of the JobResult
func (JobResult) Fields() []ent.Field {
	return []ent.Field{
		field.String("scheduled_job_id"),
		field.Enum("status").
			GoType(enums.JobExecutionStatus("")).
			Annotations(
				entgql.OrderField("status"),
			),
		field.Int("exit_code").
			Annotations(
				entgql.OrderField("exit_code"),
			).
			NonNegative().
			Nillable().
			Immutable(),

		field.Time("finished_at").
			Immutable().
			Default(time.Now).
			Annotations(
				entgql.Skip(
					entgql.SkipMutationUpdateInput,
				),
				entgql.OrderField("finished_at"),
			).
			Comment("The time the job finished it's execution. This is different from the db insertion time"),

		field.Time("started_at").
			Immutable().
			Default(time.Now).
			Annotations(
				entgql.Skip(
					entgql.SkipMutationUpdateInput,
				),
				entgql.OrderField("started_at"),
			).
			Comment("The time the job started it's execution. This is different from the db insertion time"),

		field.String("file_id").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			),
	}
}

// Mixin of the JobResult
func (j JobResult) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(j),
		},
	}.getMixins()
}

// Edges of the JobResult
func (j JobResult) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: j,
			edgeSchema: ControlScheduledJob{},
			field:      "scheduled_job_id",
			required:   true,
		}),

		uniqueEdgeTo(&edgeDefinition{
			fromSchema: j,
			edgeSchema: File{},
			field:      "file_id",
			required:   true,
		}),
	}
}

// Indexes of the JobResult
func (JobResult) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the JobResult
func (JobResult) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the JobResult
func (JobResult) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the JobResult
func (JobResult) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the JobResult
func (JobResult) Policy() ent.Policy {
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
			// this needs to be commented out for the first run that had the entfga annotation

			// entfga.CheckEditAccess[*generated.JobResultMutation](),
		),
	)
}
