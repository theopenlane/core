package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/iam/entfga"
)

// MappedControl holds the schema definition for the MappedControl entity
type MappedControl struct {
	SchemaFuncs

	ent.Schema
}

const SchemaMappedControl = "mapped_control"

func (MappedControl) Name() string {
	return SchemaMappedControl
}

func (MappedControl) GetType() any {
	return MappedControl.Type
}

func (MappedControl) PluralName() string {
	return pluralize.NewClient().Plural(SchemaMappedControl)
}

// Fields of the MappedControl
func (MappedControl) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("mapping_type").
			GoType(enums.MappingType("")).
			Comment("the type of mapping between the two controls, e.g. subset, intersect, equal, superset").
			Annotations(
				entgql.OrderField("MAPPING_TYPE"),
			).
			Default(enums.MappingTypeEqual.String()),
		field.String("relation").
			Comment("description of how the two controls are related").
			Optional(),
		field.Int("confidence").
			Comment("percentage (0-100) of confidence in the mapping").
			Min(0).
			Max(100). //nolint:mnd
			Nillable().
			Optional(),
		field.Enum("source").
			GoType(enums.MappingSource("")).
			Optional().
			Annotations(
				entgql.OrderField("SOURCE"),
			).
			Default(enums.MappingSourceManual.String()).
			Comment("source of the mapping, e.g. manual, suggested, etc."),
	}
}

// Edges of the MappedControl
func (m MappedControl) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: m,
			t:          Control.Type,
			name:       "from_controls",
			comment:    "controls that map to another control",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: m,
			t:          Control.Type,
			name:       "to_controls",
			comment:    "controls that are being mapped from another control",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: m,
			t:          Subcontrol.Type,
			name:       "from_subcontrols",
			comment:    "subcontrols map to another control",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: m,
			t:          Subcontrol.Type,
			name:       "to_subcontrols",
			comment:    "subcontrols are being mapped from another control",
		}),
	}
}

// Mixin of the MappedControl
func (m MappedControl) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(m),
			// add group edit permissions to the mapped control
			newGroupPermissionsMixin(withSkipViewPermissions()),
		},
	}.getMixins(m)
}

func (MappedControl) Features() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the MappedControl
func (m MappedControl) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Interceptors of the MappedControl
func (m MappedControl) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.InterceptorFeatures(m.Features()...),
	}
}

// Hooks of the MappedControl
func (MappedControl) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			hooks.OrgOwnedTuplesHook(),
			ent.OpCreate,
		),
	}
}

// Policy of the MappedControl
func (m MappedControl) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.DenyIfMissingAllFeatures("mappedcontrol", m.Features()...),
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.MappedControlMutation](),
		),
	)
}
