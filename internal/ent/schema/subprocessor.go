package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/ent/validator"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
)

// Subprocessor holds the schema definition for the Subprocessor entity
type Subprocessor struct {
	SchemaFuncs

	ent.Schema
}

// SchemaSubprocessor is the name of the Subprocessor schema.
const SchemaSubprocessor = "subprocessor"

// Name returns the name of the Subprocessor schema.
func (Subprocessor) Name() string {
	return SchemaSubprocessor
}

// GetType returns the type of the Subprocessor schema.
func (Subprocessor) GetType() any {
	return Subprocessor.Type
}

// PluralName returns the plural name of the Subprocessor schema.
func (Subprocessor) PluralName() string {
	return pluralize.NewClient().Plural(SchemaSubprocessor)
}

// Fields of the Subprocessor
func (Subprocessor) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			).
			Comment("name of the standard body").
			Annotations(entx.FieldSearchable()),
		field.Text("description").
			Optional().
			Comment("description of the subprocessor"),
		field.String("logo_remote_url").
			Comment("URL of the logo").
			MaxLen(urlMaxLen).
			Validate(validator.ValidateURL()).
			Optional().
			Nillable(),
		field.String("logo_local_file_id").
			Comment("The local logo file id, takes precedence over the logo remote URL").
			Optional().
			Annotations(
				// this field is not exposed to the graphql schema, it is set by the file upload handler
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			).
			Nillable(),
	}
}

// Mixin of the Subprocessor
func (t Subprocessor) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(t,
				withSkipForSystemAdmin(true), // allow empty owner_id for system admin
				withAllowAnonymousTrustCenterAccess(true),
			),
			mixin.SystemOwnedMixin{},
		},
	}.getMixins(t)
}

// Edges of the Subprocessor
func (t Subprocessor) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(t, File{}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: t,
			name:       "logo_file",
			t:          File.Type,
			field:      "logo_local_file_id",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenterSubprocessor{},
		}),
	}
}

// Hooks of the Subprocessor
func (Subprocessor) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookSubprocessor(),
	}
}

// Policy of the Subprocessor
func (s Subprocessor) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.SystemOwnedSubprocessor(),
			rule.DenyIfMissingAllFeatures("subprocessor", s.Features()...),
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Indexes of the Subprocessor
func (Subprocessor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name", "owner_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

func (Subprocessor) Features() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogTrustCenterModule,
	}
}

// Interceptors of the Subprocessor
func (t Subprocessor) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorFeatures(t.Features()...),
		interceptors.TraverseSubprocessor(),
	}
}
