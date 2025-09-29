package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"
)

// TrustCenterControl holds the schema definition for the TrustCenterControl entity
type TrustCenterControl struct {
	SchemaFuncs

	ent.Schema
}

// SchemaTrustCenterControl is the name of the TrustCenterControl schema.
const SchemaTrustCenterControl = "trust_center_control"

// Name returns the name of the TrustCenterControl schema.
func (TrustCenterControl) Name() string {
	return SchemaTrustCenterControl
}

// GetType returns the type of the TrustCenterControl schema.
func (TrustCenterControl) GetType() any {
	return TrustCenterControl.Type
}

// PluralName returns the plural name of the TrustCenterControl schema.
func (TrustCenterControl) PluralName() string {
	return pluralize.NewClient().Plural(SchemaTrustCenterControl)
}

// Fields of the TrustCenterControl
func (TrustCenterControl) Fields() []ent.Field {
	return []ent.Field{
		field.String("control_id").
			Comment("ID of the control").
			NotEmpty(),
		field.String("trust_center_id").
			Comment("ID of the trust center").
			NotEmpty().
			Optional(),
	}
}

// Mixin of the TrustCenterControl
func (t TrustCenterControl) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.TrustCenterControl](t,
				withParents(TrustCenter{}),
			),
		},
	}.getMixins(t)
}

// Edges of the TrustCenterControl
func (t TrustCenterControl) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: TrustCenter{},
			field:      "trust_center_id",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: t,
			edgeSchema: Control{},
			field:      "control_id",
			required:   true,
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Control{}.Name()),
			},
		}),
	}
}

// Hooks of the TrustCenterControl
func (TrustCenterControl) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookTrustCenterControlAuthz(),
	}
}

// Policy of the TrustCenterControl
func (TrustCenterControl) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.TrustCenterControlMutation](),
		),
	)
}

// Indexes of the TrustCenterControl
func (TrustCenterControl) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("control_id", "trust_center_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Annotations of the TrustCenterControl
func (TrustCenterControl) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SettingsChecks("trust_center"),
	}
}

// Interceptors of the TrustCenterControl
func (TrustCenterControl) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorTrustCenterChild(),
	}
}
