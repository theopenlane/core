package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/interceptors"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/entx/accessmap"
)

// ProgramMembership holds the schema definition for the ProgramMembership entity
type ProgramMembership struct {
	SchemaFuncs

	ent.Schema
}

// SchemaProgramMembership is the name of the ProgramMembership schema.
const SchemaProgramMembership = "program_membership"

// Name returns the name of the ProgramMembership schema.
func (ProgramMembership) Name() string {
	return SchemaProgramMembership
}

// GetType returns the type of the ProgramMembership schema.
func (ProgramMembership) GetType() any {
	return ProgramMembership.Type
}

// PluralName returns the plural name of the ProgramMembership schema.
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
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
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
			Unique().Annotations(),
	}
}

// Mixin of the ProgramMembership
func (ProgramMembership) Mixin() []ent.Mixin {
	return mixinConfig{excludeTags: true, excludeSoftDelete: true}.getMixins(ProgramMembership{})
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
		policy.WithMutationRules(
			entfga.CheckEditAccess[*generated.ProgramMembershipMutation](),
		),
	)
}

func (ProgramMembership) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
	}
}
