package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// DirectoryMembership associates a DirectoryAccount with a DirectoryGroup snapshot
type DirectoryMembership struct {
	SchemaFuncs

	ent.Schema
}

// SchemaDirectoryMembership is the canonical schema name
const SchemaDirectoryMembership = "directory_membership"

// Name returns the schema name
func (DirectoryMembership) Name() string {
	return SchemaDirectoryMembership
}

// GetType returns the ent type
func (DirectoryMembership) GetType() any {
	return DirectoryMembership.Type
}

// PluralName returns the pluralized schema name
func (DirectoryMembership) PluralName() string {
	return pluralize.NewClient().Plural(SchemaDirectoryMembership)
}

// Fields of the DirectoryMembership
func (DirectoryMembership) Fields() []ent.Field {
	return []ent.Field{
		field.String("integration_id").
			Comment("integration that owns this directory membership").
			NotEmpty().
			Immutable(),
		field.String("directory_sync_run_id").
			Comment("sync run that produced this snapshot").
			NotEmpty().
			Immutable(),
		field.String("directory_account_id").
			Comment("directory account participating in this membership").
			Immutable(),
		field.String("directory_group_id").
			Comment("directory group associated with this membership").
			Immutable(),
		field.Enum("role").
			Comment("membership role reported by the provider").
			GoType(enums.DirectoryMembershipRole("")).
			Default(enums.DirectoryMembershipRoleMember.String()).
			Optional(),
		field.String("source").
			Comment("mechanism used to populate the membership (api, scim, csv, etc)").
			Optional().
			Nillable(),
		field.Time("first_seen_at").
			Comment("first time the membership was detected").
			Optional().
			Nillable(),
		field.Time("last_seen_at").
			Comment("most recent time the membership was detected").
			Optional().
			Nillable(),
		field.Time("observed_at").
			Comment("time when this record was created").
			Default(time.Now).
			Immutable(),
		field.String("last_confirmed_run_id").
			Comment("sync run identifier that most recently confirmed this membership").
			Optional().
			Nillable(),
		field.JSON("metadata", map[string]any{}).
			Comment("raw metadata associated with this membership from the provider").
			Optional(),
	}
}

// Mixin of the DirectoryMembership
func (m DirectoryMembership) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:            "DRM",
		excludeTags:       true,
		excludeSoftDelete: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(m),
		},
	}.getMixins(m)
}

// Edges of the DirectoryMembership
func (m DirectoryMembership) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: m,
			edgeSchema: Integration{},
			field:      "integration_id",
			required:   true,
			immutable:  true,
			comment:    "integration that owns this directory membership",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: m,
			edgeSchema: DirectorySyncRun{},
			field:      "directory_sync_run_id",
			required:   true,
			immutable:  true,
			comment:    "sync run that produced this snapshot",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: m,
			edgeSchema: DirectoryAccount{},
			field:      "directory_account_id",
			required:   true,
			immutable:  true,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: m,
			edgeSchema: DirectoryGroup{},
			field:      "directory_group_id",
			required:   true,
			immutable:  true,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		defaultEdgeToWithPagination(m, Event{}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: m,
			edgeSchema: WorkflowObjectRef{},
			name:       "workflow_object_refs",
			ref:        "directory_membership",
		}),
	}
}

// Indexes of the DirectoryMembership
func (DirectoryMembership) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("directory_account_id", "directory_group_id", "directory_sync_run_id").
			Unique(),
		index.Fields("directory_sync_run_id"),
		index.Fields("integration_id", "directory_sync_run_id"),
	}
}

// Policy of the DirectoryMembership
func (m DirectoryMembership) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the DirectoryMembership
func (DirectoryMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.SchemaSearchable(false),
		entx.Exportable{},
	}
}
