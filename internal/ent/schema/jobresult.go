package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/entx/history"
)

// JobResult holds the schema definition for the JobResult entity
type JobResult struct {
	SchemaFuncs

	ent.Schema
}

// SchemaJobResultis the name of the schema in snake case
const SchemaJobResult = "job_result"

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
		field.String("scheduled_job_id").
			Comment("the job this result belongs to"),

		field.Enum("status").
			GoType(enums.JobExecutionStatus("")).
			Comment("the status of this job. did it fail? did it succeed?").
			Annotations(
				entgql.OrderField("STATUS"),
			),
		field.Int("exit_code").
			Annotations(
				entgql.OrderField("exit_code"),
			).
			Comment("the exit code from the script that was executed").
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
		field.Text("log").
			Comment("the log output from the job").
			Optional().
			Nillable(),
	}
}

// Mixin of the JobResult
func (j JobResult) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(j),
		},
	}.getMixins(j)
}

// Edges of the JobResult
func (j JobResult) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: j,
			edgeSchema: ScheduledJob{},
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

func (JobResult) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the JobResult
func (j JobResult) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the JobResult
func (JobResult) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookJobResultFiles(),
	}
}

// Interceptors of the JobResult
func (JobResult) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the JobResult
func (JobResult) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			// only system admin tokens should be posting back job results
			rule.AllowMutationIfSystemAdmin(),
		),
	)
}
