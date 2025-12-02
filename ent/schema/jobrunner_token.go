package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/ent/privacy/token"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/history"
	"github.com/theopenlane/shared/models"
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
			newOrgOwnedMixin(j,
				withSkipTokenTypesObjects(&token.JobRunnerRegistrationToken{})),
		},
	}.getMixins(j)
}

// Edges of the JobRunnerToken
func (j JobRunnerToken) Edges() []ent.Edge {
	return []ent.Edge{
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: j,
			edgeSchema: JobRunner{},
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Organization{}.Name()),
			},
		}),
	}
}

// Indexes of the JobRunnerToken
func (JobRunnerToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token", "expires_at", "is_active"),
	}
}

func (JobRunnerToken) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the JobRunnerToken
func (j JobRunnerToken) Annotations() []schema.Annotation {
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
func (j JobRunnerToken) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.JobRunnerRegistrationToken](),
			rule.AllowIfContextAllowRule(),
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
