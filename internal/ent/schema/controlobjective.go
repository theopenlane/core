package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// ControlObjective defines the controlobjective schema.
type ControlObjective struct {
	SchemaFuncs

	ent.Schema
}

// SchemaControlObjective is the name of the controlobjective schema.
const SchemaControlObjective = "control_objective"

// Name returns the name of the controlobjective schema.
func (ControlObjective) Name() string {
	return SchemaControlObjective
}

// GetType returns the type of the controlobjective schema.
func (ControlObjective) GetType() any {
	return ControlObjective.Type
}

// PluralName returns the plural name of the controlobjective schema.
func (ControlObjective) PluralName() string {
	return pluralize.NewClient().Plural(SchemaControlObjective)
}

// Fields returns controlobjective fields.
func (ControlObjective) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			).
			Comment("the name of the control objective"),
		field.Text("desired_outcome").
			Optional().
			Comment("the desired outcome or target of the control objective"),
		field.JSON("desired_outcome_json", []any{}).
			Optional().
			Annotations(
				entgql.Type("[Any!]"),
			).
			Comment("structured details of the control objective in JSON format"),
		field.Enum("status").
			GoType(enums.ObjectiveStatus("")).
			Optional().
			Annotations(
				entgql.OrderField("status"),
			).
			Default(enums.ObjectiveDraftStatus.String()).
			Comment("status of the control objective"),
		field.Enum("source").
			GoType(enums.ControlSource("")).
			Optional().
			Annotations(
				entgql.OrderField("SOURCE"),
			).
			Default(enums.ControlSourceUserDefined.String()).
			Comment("source of the control, e.g. framework, template, custom, etc."),
		field.String("control_objective_type").
			Optional().
			Annotations(
				entgql.OrderField("control_objective_type"),
			).
			Comment("type of the control objective e.g. compliance, financial, operational, etc."),
		field.String("category").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("category"),
			).
			Comment("category of the control"),
		field.String("subcategory").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("subcategory"),
			).
			Comment("subcategory of the control"),
	}
}

// Edges of the ControlObjective
func (c ControlObjective) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(c, Program{}),
		defaultEdgeFromWithPagination(c, Evidence{}),
		defaultEdgeFromWithPagination(c, Control{}),
		defaultEdgeFromWithPagination(c, Subcontrol{}),
		defaultEdgeFromWithPagination(c, InternalPolicy{}),
		defaultEdgeToWithPagination(c, Procedure{}),
		defaultEdgeToWithPagination(c, Risk{}),
		defaultEdgeToWithPagination(c, Narrative{}),
		defaultEdgeToWithPagination(c, Task{}),
	}
}

// Mixin of the ControlObjective
func (c ControlObjective) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:          "CLO",
		includeRevision: true,
		additionalMixins: []ent.Mixin{
			// control objectives inherit permissions from the associated programs, but must have an organization as well
			// this mixin will add the owner_id field using the OrgHook but not organization tuples are created
			// it will also create program parent tuples for the control objective when a program is associated to the control objectives
			newObjectOwnedMixin[generated.ControlObjective](c,
				withParents(Program{}, Control{}, Subcontrol{}),
				withOrganizationOwner(true),
			),
			// add groups permissions with viewer, editor, and blocked groups
			newGroupPermissionsMixin(),
			mixin.NewSystemOwnedMixin(mixin.SkipTupleCreation()),
		},
	}.getMixins(c)
}

func (ControlObjective) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogPolicyManagementAddon,
	}
}

// Annotations of the ControlObjective
func (c ControlObjective) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the ControlObjective
func (c ControlObjective) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CanCreateObjectsUnderParents([]string{
				Program{}.PluralName(),
				Control{}.PluralName(),
				Subcontrol{}.PluralName(),
			}),
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.ControlObjectiveMutation](),
		),
	)
}
