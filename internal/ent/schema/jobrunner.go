package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
)

// JobRunner holds the schema definition for the JobRunner entity
type JobRunner struct {
	SchemaFuncs

	ent.Schema
}

// SchemaJobRunneris the name of the schema in snake case
const SchemaJobRunner = "job_runner"

// Name is the name of the schema in snake case
func (JobRunner) Name() string {
	return SchemaJobRunner
}

// GetType returns the type of the schema
func (JobRunner) GetType() any {
	return JobRunner.Type
}

// PluralName returns the plural name of the schema
func (JobRunner) PluralName() string {
	return pluralize.NewClient().Plural(SchemaJobRunner)
}

// Fields of the JobRunner
func (JobRunner) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the runner").
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.Enum("status").
			GoType(enums.JobRunnerStatus("")).
			Default(enums.JobRunnerStatusOffline.String()).
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput | entgql.SkipMutationUpdateInput),
			).
			Comment("the status of this runner"),
		field.String("ip_address").
			Comment("the IP address of this runner").
			Optional(),
		field.Time("last_seen").
			Comment("the last time this runner was seen").
			Optional(),
		field.String("version").
			Comment("the version of the runner").
			Optional(),
		field.String("os").
			Comment("the operating system of the runner").
			Optional(),
		field.JSON("metadata", map[string]any{}).
			Comment("raw metadata payload for the remediation from the source system").
			Optional(),
	}
}

// Mixin of the JobRunner
func (j JobRunner) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "RUN",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(j,
				withSkipTokenTypesObjects(&token.JobRunnerRegistrationToken{}),
			),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(j)
}

// Edges of the JobRunner
func (j JobRunner) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(j, JobRunnerToken{}),
	}
}

// Indexes of the JobRunner
func (JobRunner) Indexes() []ent.Index {
	return []ent.Index{}
}

func (JobRunner) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the JobRunner
func (j JobRunner) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the JobRunner
func (JobRunner) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookJobRunnerCreate(),
		hooks.HookJobRunnerDelete(),
	}
}

// Interceptors of the JobRunner
func (j JobRunner) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorJobRunnerFilterSystemOwned(),
	}
}

// Policy of the JobRunner
func (j JobRunner) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(),
		policy.WithMutationRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.JobRunnerRegistrationToken](),
			rule.AllowIfContextAllowRule(),
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
