package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/ent/hush"
	"github.com/theopenlane/ent/mixin"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/shared/models"
)

// Hush maps configured integrations (github, slack, etc.) to organizations
type Hush struct {
	SchemaFuncs

	ent.Schema
}

// SchemaHush is the name of the Hush schema.
const SchemaHush = "secret"

// Name returns the name of the Hush schema.
func (Hush) Name() string {
	return SchemaHush
}

// GetType returns the type of the Hush schema.
func (Hush) GetType() any {
	return Hush.Type
}

// PluralName returns the plural name of the Hush schema.
func (Hush) PluralName() string {
	return pluralize.NewClient().Plural(SchemaHush)
}

// Fields of the Hush
func (Hush) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the logical name of the corresponding hush secret or it's general grouping").
			NotEmpty().
			Annotations(
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("a description of the hush value or purpose, such as github PAT").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.String("kind").
			Comment("the kind of secret, such as sshkey, certificate, api token, etc.").
			Optional().
			Annotations(
				entgql.OrderField("kind"),
			),
		field.String("secret_name").
			Comment("the generic name of a secret associated with the organization").
			Immutable().
			Optional(),
		field.String("secret_value").
			Comment("the secret value").
			Sensitive().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
				hush.EncryptField(), // Automatically encrypt this field
			).
			Optional().
			Immutable(),
		field.JSON("credential_set", models.CredentialSet{}).
			Comment("a credential set, typically where you have multiple tokens or keys that compose one credential such as when accessing s3 and using access key ID, secret key, etc.").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
				hush.EncryptField(), // Automatically encrypt this field
			),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the credential").
			Optional(),
		field.Time("last_used_at").
			Optional().
			Annotations(
				entgql.OrderField("last_used_at"),
			).
			Nillable(),
		field.Time("expires_at").
			Comment("when the token expires").
			Annotations(
				entgql.OrderField("expires_at"),
			).
			Optional().
			Nillable(),
	}
}

// Edges of the Hush
func (h Hush) Edges() []ent.Edge {
	return []ent.Edge{
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: h,
			edgeSchema: Integration{},
			comment:    "the integration associated with the secret",
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: h,
			edgeSchema: File{},
			comment:    "files associated with the integration",
		}),
		defaultEdgeToWithPagination(h, Event{}),
	}
}

// Mixin of the Hush shhhh
func (h Hush) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(h),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(h)
}

// Policy of the Hush - restricts access to organization members with write access
func (Hush) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			// restrict read access to hush secrets to organization members with write access
			policy.CheckOrgEditAccess(),
		),
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}
func (Hush) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}
