package schema

import (
	"net/mail"
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"

	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/shared/enums"
)

// DirectoryGroup mirrors group metadata from an external directory provider
type DirectoryGroup struct {
	SchemaFuncs

	ent.Schema
}

// SchemaDirectoryGroup is the canonical schema name
const SchemaDirectoryGroup = "directory_group"

// Name returns the schema name
func (DirectoryGroup) Name() string {
	return SchemaDirectoryGroup
}

// GetType returns the ent type
func (DirectoryGroup) GetType() any {
	return DirectoryGroup.Type
}

// PluralName returns the pluralized name
func (DirectoryGroup) PluralName() string {
	return pluralize.NewClient().Plural(SchemaDirectoryGroup)
}

// Fields of the DirectoryGroup
func (DirectoryGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("integration_id").
			Comment("integration that owns this directory group").
			NotEmpty().
			Immutable(),
		field.String("directory_sync_run_id").
			Comment("sync run that produced this snapshot").
			NotEmpty().
			Immutable(),
		field.String("external_id").
			Comment("stable identifier from the directory system").
			NotEmpty().
			Immutable().
			Annotations(
				entgql.OrderField("external_id"),
			),
		field.String("email").
			Comment("primary group email address, when applicable").
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("email"),
			).
			Validate(func(email string) error {
				if email == "" {
					return nil
				}
				_, err := mail.ParseAddress(email)

				return err
			}),
		field.String("display_name").
			Comment("directory supplied display name").
			Optional().
			Annotations(
				entgql.OrderField("display_name"),
			),
		field.String("description").
			Comment("free-form description captured at sync time").
			Optional().
			Nillable().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Enum("classification").
			Comment("provider classification such as security, distribution, or dynamic").
			GoType(enums.DirectoryGroupClassification("")).
			Default(enums.DirectoryGroupClassificationTeam.String()),
		field.Enum("status").
			Comment("lifecycle status reported by the directory").
			GoType(enums.DirectoryGroupStatus("")).
			Default(enums.DirectoryGroupStatusActive.String()),
		field.Bool("external_sharing_allowed").
			Comment("true when directory settings allow sharing outside the tenant").
			Optional().
			Default(false),
		field.Int("member_count").
			Comment("member count reported by the directory").
			Optional(),
		field.Time("observed_at").
			Comment("time when this snapshot was recorded").
			Default(time.Now).
			Immutable(),
		field.String("profile_hash").
			Comment("hash of the normalized payload for diffing").
			Default(""),
		field.JSON("profile", map[string]any{}).
			Comment("flattened attribute bag used for filtering/diffing").
			Optional(),
		field.String("raw_profile_file_id").
			Comment("object storage file identifier containing the raw upstream payload").
			Optional().
			Nillable().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput, entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			),
		field.String("source_version").
			Comment("cursor or ETag supplied by the source system for auditing").
			Optional().
			Nillable(),
	}
}

// Mixin of the DirectoryGroup
func (g DirectoryGroup) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:            "DRG",
		excludeSoftDelete: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(g),
		},
	}.getMixins(g)
}

// Edges of the DirectoryGroup
func (g DirectoryGroup) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: g,
			edgeSchema: Integration{},
			field:      "integration_id",
			required:   true,
			immutable:  true,
			comment:    "integration that owns this directory group",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: g,
			edgeSchema: DirectorySyncRun{},
			field:      "directory_sync_run_id",
			required:   true,
			immutable:  true,
			comment:    "sync run that produced this snapshot",
		}),
		edge.From("accounts", DirectoryAccount.Type).
			Ref("groups").
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
				entgql.RelayConnection(),
			).
			Through("members", DirectoryMembership.Type),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: DirectoryGroup{},
			edgeSchema: WorkflowObjectRef{},
			name:       "workflow_object_refs",
			ref:        "directory_group",
		}),
	}
}

// Indexes of the DirectoryGroup
func (DirectoryGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("integration_id", "external_id", "directory_sync_run_id").
			Unique(),
		index.Fields("directory_sync_run_id", "email"),
		index.Fields("integration_id", "email"),
		index.Fields(ownerFieldName, "email"),
	}
}

// Policy of the DirectoryGroup
func (g DirectoryGroup) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the DirectoryGroup
func (g DirectoryGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entx.CascadeThroughAnnotationField(
			[]entx.ThroughCleanup{
				{
					Field:   "DirectoryGroup",
					Through: "DirectoryMembership",
				},
			},
		),
	}
}
