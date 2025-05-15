package schema

import (
	"errors"
	"net"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
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
			Comment("the status of this runner"),

		field.String("ip_address").
			Immutable().
			Unique().
			Comment("the IP address of this runner").
			Validate(func(s string) error {
				ip := net.ParseIP(s)
				if ip == nil {
					return errors.New("invalid ip address") // nolint: err113
				}

				if ip.IsLoopback() {
					return errors.New("you cannot use a loopback address") // nolint: err113
				}

				if ip.IsUnspecified() {
					return errors.New("you cannot use an unspecified IP like 0.0.0.0 and others") // nolint: err113
				}

				return nil
			}),
	}
}

// Mixin of the JobRunner
func (j JobRunner) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "RUN",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(j,
				withSkipForSystemAdmin(true),
			),
			mixin.SystemOwnedMixin{},
		},
	}.getMixins()
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

// Annotations of the JobRunner
func (JobRunner) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the JobRunner
func (JobRunner) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the JobRunner
func (JobRunner) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the JobRunner
func (JobRunner) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithMutationRules(
			rule.SystemOwnedJobRunner(),
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
