package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
)

// Risk defines the risk schema.
type Risk struct {
	SchemaFuncs

	ent.Schema
}

// SchemaRisk is the name of the risk schema.
const SchemaRisk = "risk"

// Name returns the name of the risk schema.
func (Risk) Name() string {
	return SchemaRisk
}

// GetType returns the type of the risk schema.
func (Risk) GetType() any {
	return Risk.Type
}

// PluralName returns the plural name of the risk schema.
func (Risk) PluralName() string {
	return pluralize.NewClient().Plural(SchemaRisk)
}

// Fields returns risk fields.
func (Risk) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			).
			Comment("the name of the risk"),
		field.Enum("status").
			GoType(enums.RiskStatus("")).
			Default(enums.RiskOpen.String()).
			Annotations(
				entgql.OrderField("STATUS"),
			).
			Optional().
			Comment("status of the risk - open, mitigated, ongoing, in-progress, and archived."),
		field.String("risk_type").
			Annotations(
				entgql.OrderField("risk_type"),
			).
			Optional().
			Comment("type of the risk, e.g. strategic, operational, financial, external, etc."),
		field.String("category").
			Optional().
			Annotations(
				entgql.OrderField("category"),
			).
			Comment("category of the risk, e.g. human resources, operations, IT, etc."),
		field.Enum("impact").
			GoType(enums.RiskImpact("")).
			Default(enums.RiskImpactModerate.String()).
			Annotations(
				entgql.OrderField("IMPACT"),
			).
			Optional().
			Comment("impact of the risk -critical, high, medium, low"),
		field.Enum("likelihood").
			GoType(enums.RiskLikelihood("")).
			Default(enums.RiskLikelihoodMid.String()).
			Optional().
			Annotations(
				entgql.OrderField("LIKELIHOOD"),
			).
			Comment("likelihood of the risk occurring; unlikely, likely, highly likely"),
		field.Int("score").
			Optional().
			Annotations(
				entgql.OrderField("score"),
				entx.FieldSearchable(),
			).
			Comment("score of the risk based on impact and likelihood (1-4 unlikely, 5-9 likely, 10-16 highly likely, 17-20 critical)"),
		field.Text("mitigation").
			Optional().
			Comment("mitigation for the risk"),
		field.Text("details").
			Optional().
			Comment("details of the risk"),
		field.Text("business_costs").
			Annotations(
				entgql.OrderField("business_costs"),
			).
			Optional().
			Comment("business costs associated with the risk"),
		field.String("stakeholder_id").
			Optional().
			Unique().
			Comment("the id of the group responsible for risk oversight"),
		field.String("delegate_id").
			Optional().
			Unique().
			Comment("the id of the group responsible for risk oversight on behalf of the stakeholder"),
	}
}

// Edges of the Risk
func (r Risk) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeFromWithPagination(r, Control{}),
		defaultEdgeFromWithPagination(r, Subcontrol{}),
		defaultEdgeFromWithPagination(r, Procedure{}),
		defaultEdgeFromWithPagination(r, InternalPolicy{}),
		defaultEdgeFromWithPagination(r, Program{}), // risk can be associated to 1:m programs, this allow permission inheritance from the program(s)
		defaultEdgeToWithPagination(r, ActionPlan{}),
		defaultEdgeToWithPagination(r, Task{}),
		defaultEdgeToWithPagination(r, Asset{}),
		defaultEdgeToWithPagination(r, Entity{}),
		defaultEdgeToWithPagination(r, Scan{}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: r,
			name:       "stakeholder",
			t:          Group.Type,
			field:      "stakeholder_id",
			comment:    "the group of users who are responsible for risk oversight",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: r,
			name:       "delegate",
			t:          Group.Type,
			field:      "delegate_id",
			comment:    "temporary delegates for the risk, used for temporary ownership",
		}),
	}
}

// Hooks of the Risk
func (Risk) Hooks() []ent.Hook {
	return []ent.Hook{
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"stakeholder": "group",
			}, "stakeholder"),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
		hook.On(
			hooks.HookRelationTuples(map[string]string{
				"delegate": "group",
			}, "delegate"),
			ent.OpCreate|ent.OpUpdateOne|ent.OpUpdateOne,
		),
	}
}

// Mixin of the Risk
func (r Risk) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "RSK",
		additionalMixins: []ent.Mixin{
			// risks inherit permissions from the associated programs, but must have an organization as well
			// this mixin will add the owner_id field using the OrgHook but not organization tuples are created
			// it will also create program parent tuples for the risk when a program is associated to the risk
			newObjectOwnedMixin[generated.Risk](r,
				withParents(
					Program{}, Control{}, Procedure{}, ControlObjective{}, InternalPolicy{}, Subcontrol{}),
				withOrganizationOwner(true),
			),
			// add groups permissions with viewer, editor, and blocked groups
			newGroupPermissionsMixin(),
		},
	}.getMixins(r)
}

// Annotations of the Risk
func (Risk) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.Features("compliance", "policy-management", "risk-management", "asset-management", "entity-management", "continuous-compliance-automation"),
		entfga.SelfAccessChecks(),
		entx.Exportable{},
	}
}

// Policy of the Risk
func (Risk) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.RiskMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.RiskMutation](),
		),
	)
}
