package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/validator"
)

// TrustCenterEntity holds the schema definition for the TrustCenterEntity entity
type TrustCenterEntity struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustCenterEntity is the name of the schema in snake case
const SchemaTrustCenterEntity = "trust_center_entity"

// Name is the name of the schema in snake case
func (TrustCenterEntity) Name() string {
	return SchemaTrustCenterEntity
}

// GetType returns the type of the schema
func (TrustCenterEntity) GetType() any {
	return TrustCenterEntity.Type
}

// PluralName returns the plural name of the schema
func (TrustCenterEntity) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenterEntity)
}

// Fields of the TrustCenterEntity
func (TrustCenterEntity) Fields() []ent.Field {
	return []ent.Field{
		field.String("logo_file_id").
			Comment("The local logo file id").
			Optional().
			Nillable(),
		field.String("url").
			Comment("URL of customer's website").
			MaxLen(urlMaxLen).
			Validate(validator.ValidateURL()).
			Optional().
			Annotations(
				entx.FieldSearchable(),
			),
		field.String("trust_center_id").
			Immutable().
			Comment("The trust center this entity belongs to").
			NotEmpty().
			Optional(),
		field.String("name").
			Comment("The name of the tag definition").
			Immutable().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("NAME"),
			),
		field.String("entity_type_id").
			Immutable().
			Comment("The entity type for the customer entity").
			Annotations(entgql.Skip(^entgql.SkipType)).
			Optional(),
	}
}

// Mixin of the TrustCenterEntity
func (t TrustCenterEntity) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.TrustCenterEntity](t,
				withParents(TrustCenter{}),
				withAllowAnonymousTrustCenterAccess(true),
			),
			newGroupPermissionsMixin(withSkipViewPermissions()),
		},
	}.getMixins(t)
}

// Edges of the TrustCenterEntity
func (t TrustCenterEntity) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			name:       "logo_file",
			t:          File.Type,
			field:      "logo_file_id",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
			immutable:  true,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			edgeSchema: EntityType{},
			field:      "entity_type_id",
			immutable:  true,
		}),
	}
}

// Indexes of the TrustCenterEntity
func (TrustCenterEntity) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the TrustCenterEntity
func (TrustCenterEntity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
	}
}

// Hooks of the TrustCenterEntity
func (TrustCenterEntity) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTrustCenterEntityCreate(),
		hooks.HookTrustCenterEntityFiles(),
	}
}

// Interceptors of the TrustCenterEntity
func (TrustCenterEntity) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}

// Modules this schema has access to
func (TrustCenterEntity) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Policy of the TrustCenterEntity
func (TrustCenterEntity) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithOnMutationRules(ent.OpCreate,
			policy.CheckCreateAccess(),
		),
		policy.WithMutationRules(
			rule.AllowIfTrustCenterEditor(),
			policy.CanCreateObjectsUnderParents([]string{
				TrustCenter{}.Name(),
			}),
			entfga.CheckEditAccess[*generated.TrustCenterEntityMutation](),
		),
	)
}
