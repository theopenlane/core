package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/common/enums"
	"github.com/theopenlane/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// ControlImplementation holds the schema definition for the ControlImplementation entity
type ControlImplementation struct {
	SchemaFuncs

	ent.Schema
}

// SchemaImplementation is the name of the ControlImplementation schema.
const SchemaImplementation = "control_implementation"

// Name returns the name of the ControlImplementation schema.
func (ControlImplementation) Name() string {
	return SchemaImplementation
}

// GetType returns the type of the ControlImplementation schema.
func (ControlImplementation) GetType() any {
	return ControlImplementation.Type
}

// PluralName returns the plural name of the ControlImplementation schema.
func (ControlImplementation) PluralName() string {
	return pluralize.NewClient().Plural(SchemaImplementation)
}

// Fields of the ControlImplementation
func (ControlImplementation) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			GoType(enums.DocumentStatus("")).
			Default(enums.DocumentDraft.String()).
			Optional().
			Annotations(
				entgql.OrderField("STATUS"),
			).
			Comment("status of the %s, e.g. draft, published, archived, etc."),
		field.Time("implementation_date").
			Optional().
			Annotations(
				entgql.OrderField("implementation_date"),
			).
			Comment("date the control was implemented"),
		field.Bool("verified").
			Optional().
			Annotations(
				entgql.OrderField("verified"),
			).
			Comment("set to true if the control implementation has been verified"),
		field.Time("verification_date").
			Optional().
			Annotations(
				entgql.OrderField("verification_date"),
			).
			Comment("date the control implementation was verified"),
		field.Text("details").
			Optional().
			Comment("details of the control implementation"),
		field.JSON("details_json", []any{}).
			Optional().
			Annotations(
				entgql.Type("[Any!]"),
			).
			Comment("structured details of the control implementation in JSON format"),
	}
}

// Mixin of the ControlImplementation
func (c ControlImplementation) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			// subcontrols can inherit permissions from the parent control
			newObjectOwnedMixin[generated.ControlImplementation](c,
				withParents(Control{}, Subcontrol{}),
				withOrganizationOwner(true),
			),
			newGroupPermissionsMixin(),
			mixin.NewSystemOwnedMixin(),
		},
	}.getMixins(c)
}

// Edges of the ControlImplementation
func (c ControlImplementation) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(c, Control{}),
		defaultEdgeFromWithPagination(c, Subcontrol{}),

		defaultEdgeToWithPagination(c, Task{}),
	}
}

// Hooks of the ControlImplementation
func (ControlImplementation) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookControlImplementation(),
	}
}

func (ControlImplementation) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}

// Annotations of the ControlImplementation
func (c ControlImplementation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the ControlImplementation
func (c ControlImplementation) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{
				Control{}.PluralName(),
				Subcontrol{}.PluralName(),
			}),
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ControlImplementationMutation](),
		),
	)
}
