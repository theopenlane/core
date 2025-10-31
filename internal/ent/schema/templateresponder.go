package schema

import (
	"net/mail"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/history"
	"github.com/theopenlane/utils/keygen"

	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/privacy/token"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
)

// TemplateResponder holds the schema definition for the TemplateResponder entity
type TemplateResponder struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTemplateResponder is the name of the TemplateResponder schema.
const SchemaTemplateResponder = "template_responder"

// Name returns the name of the schema.
func (TemplateResponder) Name() string { return SchemaTemplateResponder }

// GetType returns the type of the schema.
func (TemplateResponder) GetType() any {
	return TemplateResponder.Type
}

// PluralName returns the plural name of the schema.
func (TemplateResponder) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTemplateResponder)
}

// Fields of the TemplateResponder
func (TemplateResponder) Fields() []ent.Field {
	return []ent.Field{
		field.String("assessment_id").
			Comment("the assement associated with this responder"),

		field.String("email").
			Comment("the email address of the recipient").
			Annotations(
				entx.FieldSearchable(),
			).
			Immutable().
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}),

		field.String("token").
			Immutable().
			Unique().
			Annotations(
				entgql.Skip(entgql.SkipAll),
			).
			DefaultFunc(func() string {
				token := keygen.PrefixedSecret("questionnaire")
				return token
			}),

		field.Bytes("secret").
			Comment("the comparison secret to verify the token's signature").
			Annotations(
				entgql.Skip(entgql.SkipAll),
			).
			NotEmpty().
			Nillable(),

		field.Int("send_attempts").
			Comment("the number of attempts made to perform email send, maximum of 5").
			Annotations(
				entgql.OrderField("send_attempts"),
				entgql.Skip(entgql.SkipMutationUpdateInput, entgql.SkipMutationCreateInput),
			).
			Default(1),

		field.Enum("status").
			Comment("the status of the template responder").
			GoType(enums.TemplateResponderStatus("")).
			Default(enums.TemplateResponderStatusPending.String()).
			Annotations(
				entgql.OrderField("STATUS"),
			),
	}
}

// Mixin of the TemplateResponder
func (t TemplateResponder) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(t,
				withSkipTokenTypesObjects(&token.TemplateResponderToken{}),
			),
		},
	}.getMixins(t)
}

// Edges of the TemplateResponder
func (t TemplateResponder) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: Assessment{},
			field:      "assessment_id",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
	}
}

// Indexes of the TemplateResponder
func (TemplateResponder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token").Unique(),
		index.Fields("assessment_id", "status"),
		index.Fields("email", ownerFieldName),
		index.Fields("email", "assessment_id").
			Unique(),
	}
}

// Annotations of the TemplateResponder
func (TemplateResponder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		history.Annotations{
			Exclude: true,
		},
		entx.Exportable{},
	}
}

// Policy of the TemplateResponder
func (TemplateResponder) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.TemplateResponderToken](),
		),
		policy.WithMutationRules(
			rule.AllowIfContextHasPrivacyTokenOfType[*token.TemplateResponderToken](),
			policy.CheckOrgWriteAccess(),
		),
	)
}

func (TemplateResponder) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Hooks of the TemplateResponder
func (TemplateResponder) Hooks() []ent.Hook {
	return []ent.Hook{}
}

func (TemplateResponder) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}
