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

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

type AudienceMember struct {
	SchemaFuncs

	ent.Schema
}

const SchemaAudienceMember = "audience_member"

func (AudienceMember) Name() string {
	return SchemaAudienceMember
}

func (AudienceMember) GetType() any {
	return AudienceMember.Type
}

func (AudienceMember) PluralName() string {
	return pluralize.NewClient().Plural(SchemaAudienceMember)
}

func (AudienceMember) Fields() []ent.Field {
	return []ent.Field{
		field.String("audience_id").
			Comment("the audience this member belongs to").
			Immutable().
			NotEmpty(),
		field.String("contact_id").
			Comment("the contact associated with the audience member").
			Optional(),
		field.String("user_id").
			Comment("the user associated with the audience member").
			Optional(),
		field.String("group_id").
			Comment("the group associated with the audience member").
			Optional(),
		field.String("subscriber_id").
			Comment("the subscriber associated with the audience member").
			Optional(),
		field.String("identity_holder_id").
			Comment("the identity holder associated with the audience member").
			Optional(),
		field.String("email").
			Comment("the email address for the audience member").
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
			Comment("the name of the audience member, if known").
			Optional().
			Annotations(
				entgql.OrderField("full_name"),
			),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the audience member").
			Optional(),
	}
}

func (a AudienceMember) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:      "AUM",
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(a),
		},
	}.getMixins(a)
}

func (a AudienceMember) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: a,
			edgeSchema: Audience{},
			field:      "audience_id",
			required:   true,
			immutable:  true,
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: a,
			edgeSchema: Contact{},
			field:      "contact_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Contact{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: a,
			edgeSchema: User{},
			field:      "user_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(User{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: a,
			edgeSchema: Group{},
			field:      "group_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Group{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: a,
			edgeSchema: Subscriber{},
			field:      "subscriber_id",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: a,
			edgeSchema: IdentityHolder{},
			field:      "identity_holder_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(IdentityHolder{}.Name()),
			},
		}),
	}
}

func (AudienceMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("audience_id", "email").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields("contact_id"),
		index.Fields("user_id"),
		index.Fields("group_id"),
		index.Fields("subscriber_id"),
		index.Fields("identity_holder_id"),
	}
}

func (AudienceMember) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogTrustCenterModule,
	}
}

func (AudienceMember) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}

func (AudienceMember) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
			policy.CheckCreateAccess(),
			policy.CanCreateObjectsUnderParents([]string{
				Audience{}.PluralName(),
			}),
		),
	)
}
