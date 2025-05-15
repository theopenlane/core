package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/entx/history"
	"github.com/theopenlane/utils/keygen"
)

// JobRunnerToken holds the schema definition for the JobRunnerToken entity
type JobRunnerToken struct {
	SchemaFuncs

	ent.Schema
}

// SchemaJobRunnerTokenis the name of the schema in snake case
const SchemaJobRunnerToken = "job_runner_token"

// Name is the name of the schema in snake case
func (JobRunnerToken) Name() string {
	return SchemaJobRunnerToken
}

// GetType returns the type of the schema
func (JobRunnerToken) GetType() any {
	return JobRunnerToken.Type
}

// PluralName returns the plural name of the schema
func (JobRunnerToken) PluralName() string {
	return pluralize.NewClient().Plural(SchemaJobRunnerToken)
}

// Fields of the JobRunnerToken
func (JobRunnerToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("job_runner_id").
			Comment("the ID of the runner this token belongs to"),
		field.String("token").
			Immutable().
			Unique().
			Annotations(
				entgql.Skip(^entgql.SkipType),
			).
			DefaultFunc(func() string {
				token := keygen.PrefixedSecret("runner")
				return token
			}),
		field.Time("expires_at").
			Comment("when the token expires").
			Annotations(
				entgql.OrderField("expires_at"),
				entgql.Skip(entgql.SkipMutationUpdateInput),
			).
			// by default tokens do not expire
			Optional().
			Nillable(),
		field.Time("last_used_at").
			Annotations(
				entgql.OrderField("last_used_at"),
			).
			Optional().
			Nillable(),
		field.Bool("is_active").
			Default(true).
			Comment("whether the token is active").
			Optional(),
		field.String("revoked_reason").
			Comment("the reason the token was revoked").
			Optional().
			Nillable(),
		field.String("revoked_by").
			Comment("the user who revoked the token").
			Optional().
			Nillable(),
		field.Time("revoked_at").
			Comment("when the token was revoked").
			Optional().
			Nillable(),
	}
}

// Mixin of the JobRunnerToken
func (j JobRunnerToken) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(j),
		},
	}.getMixins()
}

// Edges of the JobRunnerToken
func (j JobRunnerToken) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: j,
			edgeSchema: JobRunner{},
			field:      "job_runner_id",
			name:       "runner_id",
			required:   true,
		}),
	}
}

// Indexes of the JobRunnerToken
func (JobRunnerToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token", "expires_at", "is_active"),
	}
}

// Annotations of the JobRunnerToken
func (JobRunnerToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the JobRunnerToken
func (JobRunnerToken) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the JobRunnerToken
func (JobRunnerToken) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Policy of the JobRunnerToken
func (JobRunnerToken) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(),
		),
		policy.WithMutationRules(
			rule.AllowIfContextAllowRule(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
