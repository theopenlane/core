package schema

import (
	"net/mail"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// TrustCenterNDARequest holds the schema definition for the TrustCenterNDARequest entity
type TrustCenterNDARequest struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustCenterNDARequest is the name of the TrustCenterNDARequest schema.
const SchemaTrustCenterNDARequest = "trust_center_nda_request"

// Name returns the name of the TrustCenterNDARequest schema.
func (TrustCenterNDARequest) Name() string {
	return SchemaTrustCenterNDARequest
}

// GetType returns the type of the TrustCenterNDARequest schema.
func (TrustCenterNDARequest) GetType() any {
	return TrustCenterNDARequest.Type
}

// PluralName returns the plural name of the TrustCenterNDARequest schema.
func (TrustCenterNDARequest) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenterNDARequest)
}

// Fields of the TrustCenterNDARequest
func (TrustCenterNDARequest) Fields() []ent.Field {
	return []ent.Field{
		field.String("trust_center_id").
			Comment("ID of the trust center").
			NotEmpty().
			Immutable().
			Optional(),
		field.String("first_name").
			Comment("first name of the requester").
			NotEmpty(),
		field.String("last_name").
			Comment("last name of the requester").
			NotEmpty(),
		field.String("email").
			Comment("email address of the requester").
			NotEmpty().
			Validate(func(email string) error {
				_, err := mail.ParseAddress(email)
				return err
			}),
		field.String("company_name").
			Comment("company name of the requester").
			Optional().
			Nillable(),
		field.String("reason").
			Comment("reason for the NDA request").
			Optional().
			Nillable(),
		field.Enum("access_level").
			GoType(enums.TrustCenterNDARequestAccessLevel("")).
			Default(enums.TrustCenterNDARequestAccessLevelFull.String()).
			Optional().
			Comment("access level requested"),
		field.Enum("status").
			GoType(enums.TrustCenterNDARequestStatus("")).
			Default(enums.TrustCenterNDARequestStatusRequested.String()).
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput),
			).
			Comment("status of the NDA request"),
		field.Time("approved_at").
			Comment("timestamp when the request was approved").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.String("approved_by_user_id").
			Comment("ID of the user who approved the request").
			Optional().
			Nillable(),
		field.Time("signed_at").
			Comment("timestamp when the NDA was signed").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.String("document_data_id").
			Comment("ID of the signed NDA document data").
			Optional().
			Nillable(),
		field.String("file_id").
			Comment("ID of the template file at the time the NDA was signed").
			Optional().
			Nillable(),
	}
}

// Mixin of the TrustCenterNDARequest
func (t TrustCenterNDARequest) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.TrustCenterNDARequest](t,
				withParents(TrustCenter{}),
				withAllowAnonymousTrustCenterAccess(true),
			),
			newGroupPermissionsMixin(withSkipViewPermissions()),
		},
	}.getMixins(t)
}

// Edges of the TrustCenterNDARequest
func (t TrustCenterNDARequest) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
			immutable:  true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			name:       "documents_requested",
			edgeSchema: TrustCenterDoc{},
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(TrustCenterDoc{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			edgeSchema: DocumentData{},
			field:      "document_data_id",
			comment:    "the signed NDA document data",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			edgeSchema: File{},
			field:      "file_id",
			comment:    "the template file at the time the NDA was signed",
		}),
	}
}

// Interceptors of the TrustCenterNDARequest
func (TrustCenterNDARequest) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}

// Hooks of the TrustCenterNDARequest
func (TrustCenterNDARequest) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTrustCenterNDARequestCreate(),
		hooks.HookTrustCenterNDARequestUpdate(),
	}
}

// Policy of the TrustCenterNDARequest
func (TrustCenterNDARequest) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowIfTrustCenterEditor(),
			policy.CheckCreateAccess(),
			policy.CanCreateObjectsUnderParents([]string{
				TrustCenter{}.Name(),
			}),
			entfga.CheckEditAccess[*generated.TrustCenterNDARequestMutation](),
		),
	)
}

func (TrustCenterNDARequest) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Indexes of the TrustCenterNDARequest
func (TrustCenterNDARequest) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the TrustCenterNDARequest
func (TrustCenterNDARequest) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
		entfga.SelfAccessChecks(),
	}
}
