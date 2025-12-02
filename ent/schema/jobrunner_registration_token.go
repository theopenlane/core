package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/ent/hooks"
	"github.com/theopenlane/ent/interceptors"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/ent/privacy/token"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/history"
	"github.com/theopenlane/utils/keygen"
)

const (
	defaultRunnerRegistrationTokenExpiration = time.Hour * 24
)

// JobRunnerRegistrationToken holds the schema definition for the JobRunnerRegistrationToken entity
type JobRunnerRegistrationToken struct {
	SchemaFuncs

	ent.Schema
}

// SchemaJobRunnerRegistrationTokenis the name of the schema in snake case
const SchemaJobRunnerRegistrationToken = "job_runner_registration_token"

// Name is the name of the schema in snake case
func (JobRunnerRegistrationToken) Name() string {
	return SchemaJobRunnerRegistrationToken
}

// GetType returns the type of the schema
func (JobRunnerRegistrationToken) GetType() any {
	return JobRunnerRegistrationToken.Type
}

// PluralName returns the plural name of the schema
func (JobRunnerRegistrationToken) PluralName() string {
	return pluralize.NewClient().Plural(SchemaJobRunnerRegistrationToken)
}

// Fields of the JobRunnerRegistrationToken
func (JobRunnerRegistrationToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("token").
			Immutable().
			Unique().
			Annotations(
				entgql.Skip(^entgql.SkipType),
			).
			DefaultFunc(func() string {
				token := keygen.PrefixedSecret("registration")
				return token
			}),
		field.Time("expires_at").
			Comment("when the token expires").
			Annotations(
				entgql.OrderField("expires_at"),
				entgql.Skip(^entgql.SkipType),
			).
			Default(time.Now().Add(defaultRunnerRegistrationTokenExpiration)).
			Immutable(),
		field.Time("last_used_at").
			Annotations(
				entgql.OrderField("last_used_at"),
			).
			Optional().
			Nillable(),
		field.String("job_runner_id").
			Optional().
			// if not optional and set, then this has been used to register a runner
			// and cannot be used again
			// Ideally, it will be deleted anyways once used
			Comment("the ID of the runner this token was used to register"),
	}
}

// Mixin of the JobRunnerRegistrationToken
func (j JobRunnerRegistrationToken) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(j, withSkipTokenTypesObjects(&token.JobRunnerRegistrationToken{})),
		},
	}.getMixins(j)
}

// Edges of the JobRunnerRegistrationToken
func (j JobRunnerRegistrationToken) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: j,
			edgeSchema: JobRunner{},
			field:      "job_runner_id",
			name:       "runner_id",
			required:   false,
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Organization{}.Name()),
			},
		}),
	}
}

// Indexes of the JobRunnerRegistrationToken
func (JobRunnerRegistrationToken) Indexes() []ent.Index {
	return []ent.Index{}
}

func (JobRunnerRegistrationToken) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the JobRunnerRegistrationToken
func (j JobRunnerRegistrationToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
	}
}

// Hooks of the JobRunnerRegistrationToken
func (JobRunnerRegistrationToken) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookJobRunnerRegistrationToken(),
	}
}

// Interceptors of the JobRunnerRegistrationToken
func (j JobRunnerRegistrationToken) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorJobRunnerRegistrationToken(),
	}
}

// Policy of the JobRunnerRegistrationToken
func (j JobRunnerRegistrationToken) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.JobRunnerRegistrationToken](),
			rule.AllowIfContextAllowRule(),
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
		),
	)
}
