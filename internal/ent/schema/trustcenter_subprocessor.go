package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/iam/entfga"
)

const (
	trustCenterSubprocessorCategoryMaxLen = 255
)

// TrustCenterSubprocessor holds the schema definition for the TrustCenterSubprocessor entity
type TrustCenterSubprocessor struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustCenterSubprocessor is the name of the TrustCenterSubprocessor schema.
const SchemaTrustCenterSubprocessor = "trust_center_subprocessor"

// Name returns the name of the TrustCenterSubprocessor schema.
func (TrustCenterSubprocessor) Name() string {
	return SchemaTrustCenterSubprocessor
}

// GetType returns the type of the TrustCenterSubprocessor schema.
func (TrustCenterSubprocessor) GetType() any {
	return TrustCenterSubprocessor.Type
}

// PluralName returns the plural name of the TrustCenterSubprocessor schema.
func (TrustCenterSubprocessor) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenterSubprocessor)
}

// Fields of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Fields() []ent.Field {
	return []ent.Field{
		field.String("subprocessor_id").
			Comment("ID of the subprocessor").
			NotEmpty(),
		field.String("trust_center_id").
			Comment("ID of the trust center").
			NotEmpty().
			Optional(),
		field.JSON("countries", []string{}).
			Comment("country codes or country where the subprocessor is located").
			Optional(),
		field.String("category").
			Comment("Category of the subprocessor, e.g. 'Data Warehouse' or 'Infrastructure Hosting'").
			NotEmpty().
			MaxLen(trustCenterSubprocessorCategoryMaxLen),
	}
}

// Mixin of the TrustCenterSubprocessor
func (t TrustCenterSubprocessor) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
	}.getMixins(t)
}

// Edges of the TrustCenterSubprocessor
func (t TrustCenterSubprocessor) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: Subprocessor{},
			field:      "subprocessor_id",
			required:   true,
		}),
	}
}

// Hooks of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Policy of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.TrustCenterSubprocessorMutation](),
		),
	)
}

// Indexes of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("subprocessor_id", "trust_center_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
	}
}

// Interceptors of the TrustCenterSubprocessor
func (TrustCenterSubprocessor) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}
