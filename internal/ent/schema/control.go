package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Control defines the control schema.
type Control struct {
	SchemaFuncs

	ent.Schema
}

// SchemaControl is the name of the control schema.
const SchemaControl = "control"

// Name returns the name of the control schema.
func (Control) Name() string {
	return SchemaControl
}

// GetType returns the type of the control schema.
func (Control) GetType() any {
	return Control.Type
}

// PluralName returns the plural name of the control schema.
func (Control) PluralName() string {
	return pluralize.NewClient().Plural(SchemaControl)
}

// Fields returns control fields.
func (Control) Fields() []ent.Field {
	// add any fields that are specific to the parent control here
	additionalFields := []ent.Field{
		field.String("ref_code").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("ref_code"),
			).
			Comment("the unique reference code for the control"),
		field.String("standard_id").
			Comment("the id of the standard that the control belongs to, if applicable").
			Optional(),
	}

	return additionalFields
}

// Edges of the Control
func (c Control) Edges() []ent.Edge {
	return []ent.Edge{
		// parents of the control (standard, program)
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Standard{},
			field:      "standard_id",
		}),
		defaultEdgeFromWithPagination(c, Program{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: ControlImplementation{},
			comment:    "the implementation(s) of the control",
		}),
		// controls have control objectives and subcontrols
		edgeToWithPagination(&edgeDefinition{
			fromSchema:    c,
			edgeSchema:    Subcontrol{},
			cascadeDelete: "Control",
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: ControlScheduledJob{},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			ref:        "to_controls",
			name:       "mapped_to_controls",
			t:          MappedControl.Type,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: c,
			ref:        "from_controls",
			name:       "mapped_from_controls",
			t:          MappedControl.Type,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
	}
}

func (Control) Indexes() []ent.Index {
	return []ent.Index{
		// ref_code should be unique within the standard
		index.Fields("standard_id", "ref_code").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL AND owner_id is NULL"),
		),
	}
}

// Mixin of the Control
func (c Control) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "CTL",
		additionalMixins: []ent.Mixin{
			// add the common overlap between control and subcontrol
			ControlMixin{
				SchemaType: c,
			},
			// controls must be associated with an organization but do not inherit permissions from the organization
			// controls can inherit permissions from the associated programs
			newObjectOwnedMixin[generated.Control](c,
				withParents(Organization{}, Program{}, Standard{}),
				withOrganizationOwner(true),
			),
			// add groups permissions with editor, and blocked groups
			// skip view because controls are automatically viewable by all users in the organization
			newGroupPermissionsMixin(withSkipViewPermissions()),
		},
	}.getMixins()
}

// Policy of the Control
func (Control) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.ControlMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ControlMutation](),
		),
	)
}
