package schema

import (
	"net/mail"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
)

// AudienceMember holds the schema definition for manually managed audience recipients.
type AudienceMember struct {
	SchemaFuncs

	ent.Schema
}

// SchemaAudienceMember is the name of the AudienceMember schema.
const SchemaAudienceMember = "audience_member"

// Name returns the name of the AudienceMember schema.
func (AudienceMember) Name() string {
	return SchemaAudienceMember
}

// GetType returns the type of the AudienceMember schema.
func (AudienceMember) GetType() any {
	return AudienceMember.Type
}

// PluralName returns the plural name of the AudienceMember schema.
func (AudienceMember) PluralName() string {
	return pluralize.NewClient().Plural(SchemaAudienceMember)
}

// Fields of the AudienceMember.
func (AudienceMember) Fields() []ent.Field {
	return []ent.Field{
		field.String("audience_id").
			Comment("the audience this member belongs to").
			Immutable().
			NotEmpty(),
		field.String("contact_id").
			Comment("the contact associated with this audience member").
			Optional(),
		field.String("user_id").
			Comment("the user associated with this audience member").
			Optional(),
		field.String("group_id").
			Comment("the group associated with this audience member").
			Optional(),
		field.String("identity_holder_id").
			Comment("the identity holder associated with this audience member").
			Optional(),
		field.String("subscriber_id").
			Comment("the subscriber associated with this audience member").
			Optional(),
		field.String("email").
			Comment("the email address for this audience member").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("email"),
			).
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}),
		field.String("full_name").
			Comment("the name of this audience member, if known").
			Optional().
			Annotations(
				entgql.OrderField("full_name"),
			),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the audience member").
			Optional(),
	}
}

// Mixin of the AudienceMember.
func (m AudienceMember) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "AUDM",
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(m),
		},
	}.getMixins(m)
}

// Edges of the AudienceMember.
func (m AudienceMember) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: m,
			edgeSchema: Audience{},
			field:      "audience_id",
			required:   true,
			immutable:  true,
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: m,
			edgeSchema: Contact{},
			field:      "contact_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Contact{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: m,
			edgeSchema: User{},
			field:      "user_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(User{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: m,
			edgeSchema: Group{},
			field:      "group_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Group{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: m,
			edgeSchema: IdentityHolder{},
			field:      "identity_holder_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(IdentityHolder{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: m,
			edgeSchema: Subscriber{},
			field:      "subscriber_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Organization{}.Name()),
			},
		}),
	}
}

// Indexes of the AudienceMember.
func (AudienceMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("audience_id", "email").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Modules this schema has access to.
func (AudienceMember) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogTrustCenterModule,
	}
}

// Annotations of the AudienceMember.
func (AudienceMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.NewExportable(),
	}
}

// Policy of the AudienceMember.
func (AudienceMember) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CanCreateObjectsUnderParents([]string{
				Audience{}.PluralName(),
			}),
			entfga.CheckEditAccess[*generated.AudienceMemberMutation](),
		),
	)
}
