package schema

import (
	"net/mail"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/utils/rout"
)

// IdentityHolder holds the schema definition for the IdentityHolder entity
type IdentityHolder struct {
	SchemaFuncs

	ent.Schema
}

// SchemaIdentityHolder is the name of the IdentityHolder schema
const SchemaIdentityHolder = "identity_holder"

// Name returns the name of the IdentityHolder schema
func (IdentityHolder) Name() string {
	return SchemaIdentityHolder
}

// GetType returns the type of the IdentityHolder schema
func (IdentityHolder) GetType() any {
	return IdentityHolder.Type
}

// PluralName returns the plural name of the IdentityHolder schema
func (IdentityHolder) PluralName() string {
	return "identity_holders"
}

// Fields of the IdentityHolder
func (IdentityHolder) Fields() []ent.Field {
	return []ent.Field{
		field.String("full_name").
			Comment("the full name of the identity holder").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("full_name"),
			),
		field.String("email").
			Comment("the email address of the identity holder").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("email"),
			).
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}),
		field.String("alternate_email").
			Comment("alternate email address for the identity holder").
			Optional().
			Annotations(
				entgql.OrderField("alternate_email"),
			).
			Validate(func(email string) error {
				if email == "" {
					return nil
				}
				_, err := mail.ParseAddress(email)
				return err
			}),
		field.String("phone_number").
			Comment("phone number for the identity holder").
			Validate(func(s string) error {
				if s == "" {
					return nil
				}

				valid := validator.ValidatePhoneNumber(s)
				if !valid {
					return rout.InvalidField("phone_number")
				}

				return nil
			}).
			Optional(),
		field.Bool("is_openlane_user").
			Comment("whether the identity holder record is linked to an Openlane user account").
			Default(false).
			Optional().
			Annotations(
				entgql.OrderField("is_openlane_user"),
			),
		field.String("user_id").
			Comment("the user id associated with the identity holder record").
			Optional(),
		field.Enum("identity_holder_type").
			Comment("the classification of identity holders, such as employee or contractor").
			GoType(enums.IdentityHolderType("")).
			Default(enums.IdentityHolderTypeEmployee.String()).
			Annotations(
				entgql.OrderField("IDENTITY_HOLDER_TYPE"),
			),
		field.Enum("status").
			Comment("the status of the identity holder record").
			GoType(enums.UserStatus("")).
			Default(enums.UserStatusActive.String()).
			Annotations(
				entgql.OrderField("STATUS"),
				entx.FieldWorkflowEligible(),
			),
		field.Bool("is_active").
			Comment("whether the identity holder record is active").
			Default(true).
			Annotations(
				entgql.OrderField("is_active"),
				entx.FieldWorkflowEligible(),
			),
		field.String("title").
			Comment("the job title of the identity holder").
			Optional().
			Annotations(
				entgql.OrderField("title"),
			),
		field.String("department").
			Comment("the department or function of the identity holder").
			Optional().
			Annotations(
				entgql.OrderField("department"),
			),
		field.String("team").
			Comment("the team name for the identity holder").
			Optional().
			Annotations(
				entgql.OrderField("team"),
			),
		field.String("location").
			Comment("location or office for the identity holder").
			Optional().
			Annotations(
				entgql.OrderField("location"),
			),
		field.Time("start_date").
			Comment("the start date for the identity holder").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("start_date"),
				entx.FieldWorkflowEligible(),
			),
		field.Time("end_date").
			Comment("the end date for the identity holder, if applicable").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("end_date"),
				entx.FieldWorkflowEligible(),
			),
		field.String("employer_entity_id").
			Comment("the external entity this identity holder is affiliated with").
			Optional(),
		field.String("external_user_id").
			Comment("external user identifier for the identity holder").
			Optional().
			Annotations(
				entgql.OrderField("external_user_id"),
			),
		field.String("external_reference_id").
			Comment("external identifier for the identity holder from an upstream roster").
			Optional().
			Annotations(
				entgql.OrderField("external_reference_id"),
			),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the identity holder").
			Optional(),
	}
}

// Mixin of the IdentityHolder
func (p IdentityHolder) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "IDH",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[IdentityHolder](p,
				withParents(Organization{}, Platform{}),
				withOrganizationOwner(true),
			),
			//			newGroupPermissionsMixin(),
			newResponsibilityMixin(p, withInternalOwner()),
			newCustomEnumMixin(p, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(p, withEnumFieldName("scope"), withGlobalEnum()),
			WorkflowApprovalMixin{},
		},
	}.getMixins(p)
}

// Edges of the IdentityHolder
func (p IdentityHolder) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: p,
			name:       "employer",
			t:          Entity.Type,
			field:      "employer_entity_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Entity{}.Name()),
			},
		}),
		defaultEdgeToWithPagination(p, AssessmentResponse{}),
		defaultEdgeToWithPagination(p, Assessment{}),
		defaultEdgeToWithPagination(p, Template{}),
		defaultEdgeToWithPagination(p, Asset{}),
		defaultEdgeToWithPagination(p, Entity{}),
		defaultEdgeFromWithPagination(p, Platform{}),
		defaultEdgeFromWithPagination(p, Campaign{}),
		defaultEdgeToWithPagination(p, Task{}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: p,
			edgeSchema: WorkflowObjectRef{},
			name:       "workflow_object_refs",
			ref:        "identity_holder",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: p,
			t:          Platform.Type,
			name:       "access_platforms",
			comment:    "platforms the identity holder has access to",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: p,
			name:       "user",
			t:          User.Type,
			field:      "user_id",
			ref:        "identity_holder_profiles",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(User{}.Name()),
			},
		}),
	}
}

// Indexes of the IdentityHolder
func (IdentityHolder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
		index.Fields("user_id"),
		index.Fields("external_user_id"),
	}
}

// Modules this schema has access to
func (IdentityHolder) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the IdentityHolder
func (IdentityHolder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the IdentityHolder
func (IdentityHolder) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CanCreateObjectsUnderParents([]string{
				Platform{}.PluralName(),
			}),
			//			entfga.CheckEditAccess[*generated.IdentityHolderMutation](),
		),
	)
}
