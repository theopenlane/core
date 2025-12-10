package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/graphapi/directives"
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
				directives.ExternalSourceDirectiveAnnotation,
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
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: ScheduledJob{},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: s,
			ref:        "to_subcontrols",
			name:       "mapped_to_subcontrols",
			t:          MappedControl.Type,
			annotations: []schema.Annotation{
				entgql.Skip(^entgql.SkipWhereInput),
			},
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: s,
			ref:        "from_subcontrols",
			name:       "mapped_from_subcontrols",
			t:          MappedControl.Type,
			annotations: []schema.Annotation{
				entgql.Skip(^entgql.SkipWhereInput),
			},
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
			mixin.NewSystemOwnedMixin(),
			newCustomEnumMixin(s),
		},
	}.getMixins(s)
}

// Indexes of the Subcontrol
func (Subcontrol) Indexes() []ent.Index {
	return []ent.Index{
		// ref_code should be unique within the parent control
		index.Fields("control_id", "ref_code").
			Unique().Annotations(
			entsql.IndexWhere("deleted_at is NULL"),
		),
		index.Fields("control_id", "ref_code", "owner_id").
			Annotations(
				entsql.IndexWhere("deleted_at is NULL"),
			),
		index.Fields("reference_id", "deleted_at", "owner_id"),
		index.Fields("auditor_reference_id", "deleted_at", "owner_id"),
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
func (s Subcontrol) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowIfContextAllowRule(),
			policy.CanCreateObjectsUnderParents([]string{
				Control{}.Name(),
			}),
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.SubcontrolMutation](),
		),
	)
}

// Annotations of the Standard
func (Subcontrol) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}
