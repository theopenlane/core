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
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
)

// Narrative defines the narrative schema
type Narrative struct {
	SchemaFuncs

	ent.Schema
}

const SchemaNarrative = "narrative"

func (Narrative) Name() string {
	return SchemaNarrative
}

func (Narrative) GetType() any {
	return Narrative.Type
}

func (Narrative) PluralName() string {
	return pluralize.NewClient().Plural(SchemaNarrative)
}

// Fields returns narrative fields
func (Narrative) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			).
			Comment("the name of the narrative"),
		field.Text("description").
			Optional().
			Annotations(
				entx.FieldSearchable(),
			).
			Comment("the description of the narrative"),
		field.Text("details").
			Optional().
			Comment("text data for the narrative document"),
	}
}

// Edges of the Narrative
func (n Narrative) Edges() []ent.Edge {
	return []ent.Edge{
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: n,
			name:       "satisfies",
			t:          Control.Type,
			comment:    "which controls are satisfied by the narrative",
		}),
		defaultEdgeFromWithPagination(n, Program{}),
	}
}

// Mixin of the Narrative
func (n Narrative) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "NRT",
		additionalMixins: []ent.Mixin{
			// narratives inherit permissions from the associated programs, but must have an organization as well
			// this mixin will add the owner_id field using the OrgHook but not organization tuples are created
			// it will also create program parent tuples for the narrative when a program is associated to the narrative
			newObjectOwnedMixin(n,
				withParents(Program{}),
				withOrganizationOwner(false),
			),
			// add groups permissions with viewer, editor, and blocked groups
			newGroupPermissionsMixin(),
		},
	}.getMixins()
}

// Annotations of the Narrative
func (Narrative) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the Narrative
func (Narrative) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results

		),
		policy.WithMutationRules(
			rule.CanCreateObjectsUnderParent[*generated.NarrativeMutation](rule.ProgramParent), // if mutation contains program_id, check access
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.NarrativeMutation](),
		),
	)
}
