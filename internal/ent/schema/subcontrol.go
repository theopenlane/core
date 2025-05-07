package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Subcontrol defines the file schema.
type Subcontrol struct {
	SchemaFuncs

	ent.Schema
}

// SchemaSubcontrol is the name of the Subcontrol schema.
const SchemaSubcontrol = "subcontrol"

// Name returns the name of the Subcontrol schema.
func (Subcontrol) Name() string {
	return SchemaSubcontrol
}

// GetType returns the type of the Subcontrol schema.
func (Subcontrol) GetType() any {
	return Subcontrol.Type
}

// PluralName returns the plural name of the Subcontrol schema.
func (Subcontrol) PluralName() string {
	return pluralize.NewClient().Plural(SchemaSubcontrol)
}

// Fields returns file fields.
func (Subcontrol) Fields() []ent.Field {
	// add any fields that are specific to the subcontrol here
	additionalFields := []ent.Field{
		field.String("ref_code").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("ref_code"),
			).
			Comment("the unique reference code for the control"),
		field.String("control_id").
			Unique().
			Comment("the id of the parent control").
			NotEmpty(),
	}

	return additionalFields
}

// Edges of the Subcontrol
func (s Subcontrol) Edges() []ent.Edge {
	return []ent.Edge{
		// subcontrols are required to have a parent control
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: s,
			edgeSchema: Control{},
			field:      "control_id",
			required:   true,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: ControlImplementation{},
			comment:    "the implementation(s) of the subcontrol",
		}),
	}
}

// Mixin of the Subcontrol
func (s Subcontrol) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "SCL",
		additionalMixins: []ent.Mixin{
			// add the common overlap between control and subcontrol
			ControlMixin{
				SchemaType: s,
			},
			// subcontrols can inherit permissions from the parent control
			newObjectOwnedMixin[generated.Subcontrol](s,
				withParents(Control{}),
				withOrganizationOwner(true),
			),
		},
	}.getMixins()
}

// Indexes of the Subcontrol
func (Subcontrol) Indexes() []ent.Index {
	return []ent.Index{
		// ref_code should be unique within the parent control
		index.Fields("control_id", "ref_code").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL"),
		),
	}
}

// Hooks of the Subcontrol
func (Subcontrol) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookSubcontrolCreate(),
		hooks.HookSubcontrolUpdate(),
	}
}

// Policy of the Subcontrol
func (Subcontrol) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.SubcontrolMutation](rule.ControlParent), // if mutation contains control_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.SubcontrolMutation](),
		),
	)
}
