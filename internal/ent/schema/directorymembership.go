package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/entx/history"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/privacy/policy"
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
			Annotations(
				entx.IntegrationMappingField().SystemControlled(),
			),
		field.String("platform_id").
			Comment("optional platform associated with this directory membership").
			Optional().
			NotEmpty().
			Immutable(),
		field.String("directory_account_id").
			Comment("directory account participating in this membership").
			NotEmpty().
			Immutable().
			Annotations(
				entx.IntegrationMappingField().LookupKey(),
			),
		field.String("directory_group_id").
			Comment("directory group associated with this membership").
			NotEmpty().
			Immutable().
			Annotations(
				entx.IntegrationMappingField().LookupKey(),
			),
		field.Enum("role").
			Comment("membership role reported by the provider").
			GoType(enums.DirectoryMembershipRole("")).
			Default(enums.DirectoryMembershipRoleMember.String()).
			Optional(),
		field.String("source").
			Comment("mechanism used to populate the membership (api, scim, csv, etc)").
			Optional().
			Nillable(),
		field.String("directory_name").
			Comment("directory source label set by the integration (e.g. googleworkspace, github, slack)").
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("directory_name"),
			),
		field.Time("added_at").
			Comment("provider-reported time the membership was added in the source directory").
			Optional().
			Nillable().
			Annotations(
				entx.IntegrationMappingField(),
			),
		field.Time("removed_at").
			Comment("provider-reported or locally-recorded time the membership was removed from the source directory").
			Optional().
			Nillable().
			Annotations(
				entx.IntegrationMappingField(),
				entx.SnapshotRemoval().Episodic(),
			),
		field.Time("observed_at").
			Comment("time when this record was created").
			Default(time.Now).
			Immutable(),
		field.JSON("metadata", map[string]any{}).
			Comment("raw metadata associated with this membership from the provider").
			Optional().
			Annotations(
				entx.IntegrationMappingField().Volatile(),
			),
	}
}

// Mixin of the DirectoryMembership
func (m DirectoryMembership) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:            "DRM",
		excludeTags:       true,
		excludeSoftDelete: true,
		additionalMixins: []ent.Mixin{
			ProvenanceMixin{SchemaType: m},
			newOrgOwnedMixin(m),
			newCustomEnumMixin(m, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(m, withEnumFieldName("scope"), withGlobalEnum()),
		},
	}.getMixins(m)
}

// Edges of the DirectoryMembership
func (m DirectoryMembership) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: m,
			edgeSchema: Integration{},
			field:      "integration_id",
			required:   true,
			comment:    "integration that owns this directory membership",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Organization{}.Name()),
			},
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: m,
			edgeSchema: Platform{},
			field:      "platform_id",
			immutable:  true,
			comment:    "platform associated with this directory membership",
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
		// declaring the pair index ourselves keeps ent from adding its own fully-unique version
		// for the M2M through edges; the partial predicate limits uniqueness to active rows so
		// removed membership episodes can accumulate per (account, group) pair
		index.Fields("directory_account_id", "directory_group_id").
			Unique().
			Annotations(entsql.IndexWhere("removed_at is NULL")),
		index.Fields("owner_id", "managed_by", "source_definition_id", "source_instance_id").
			Annotations(entsql.IndexWhere("removed_at is NULL")),
	}
}

// Policy of the DirectoryMembership
func (m DirectoryMembership) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
			policy.CheckCreateAccess(),
		),
	)
}

// Annotations of the DirectoryMembership
func (DirectoryMembership) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.SchemaSearchable(false),
		entx.NewExportable(),
		entx.IntegrationMappingSchema().StockPersist().InstanceScoped(),
		history.Annotations{
			Exclude: true,
		},
	}
}
