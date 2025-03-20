package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/pkg/enums"
)

// ProgramMembership holds the schema definition for the ProgramMembership entity
type ProgramMembership struct {
	SchemaFuncs

	ent.Schema
}

const SchemaProgramMembership = "programmembership"

func (ProgramMembership) Name() string {
	return SchemaProgramMembership
}

func (ProgramMembership) GetType() any {
	return ProgramMembership.Type
}

func (ProgramMembership) PluralName() string {
	return pluralize.NewClient().Plural(SchemaProgramMembership)
}

// Fields of the ProgramMembership
func (ProgramMembership) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("role").
			GoType(enums.Role("")).
			Default(string(enums.RoleMember)).
			Annotations(
				entgql.OrderField("ROLE"),
			),
		field.String("program_id").Immutable(),
		field.String("user_id").Immutable(),
	}
}

// Edges of the ProgramMembership
func (p ProgramMembership) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: p,
			edgeSchema: Program{},
			required:   true,
			immutable:  true,
			field:      "program_id",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: p,
			edgeSchema: User{},
			required:   true,
			immutable:  true,
			field:      "user_id",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: p,
			edgeSchema: OrgMembership{}, // ensure this isn't messed up
			immutable:  true,
			annotations: []schema.Annotation{
				entgql.Skip(entgql.SkipAll),
			},
		}),
	}
}

// Annotations of the ProgramMembership
func (ProgramMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.MembershipChecks("program"),
	}
}

// Indexes of the ProgramMembership
func (ProgramMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "program_id").
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Mixin of the ProgramMembership
func (ProgramMembership) Mixin() []ent.Mixin {
	return mixinConfig{excludeTags: true}.getMixins()
}

// Hooks of the ProgramMembership
func (ProgramMembership) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookProgramMembers(),
		hooks.HookMembershipSelf("program_memberships"),
	}
}

// Interceptors of the ProgramMembership
func (ProgramMembership) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{
		interceptors.FilterListQuery(),
	}
}

// // Policy of the ProgramMembership
func (ProgramMembership) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			privacy.AlwaysAllowRule(), //  interceptor should filter out the results
		),
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.ProgramMembershipMutation](),
		),
	)
}
