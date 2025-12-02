package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/ent/privacy/policy"
)

// DirectoryAccount captures one normalized identity fetched from an external directory provider
type DirectoryAccount struct {
	SchemaFuncs

	ent.Schema
}

// SchemaDirectoryAccount is the canonical schema name
const SchemaDirectoryAccount = "directory_account"

// Name returns the schema name.
func (DirectoryAccount) Name() string {
	return SchemaDirectoryAccount
}

// GetType returns the ent type for DirectoryAccount
func (DirectoryAccount) GetType() any {
	return DirectoryAccount.Type
}

// PluralName returns the pluralized schema name
func (DirectoryAccount) PluralName() string {
	return pluralize.NewClient().Plural(SchemaDirectoryAccount)
}

// Fields of the DirectoryAccount
func (DirectoryAccount) Fields() []ent.Field {
	return []ent.Field{
		field.String("integration_id").
			Comment("integration that owns this directory account").
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
		field.String("secondary_key").
			Comment("optional secondary identifier such as Azure immutable ID").
			Optional().
			Nillable(),
		field.String("canonical_email").
			Comment("lower-cased primary email address, if present").
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("canonical_email"),
			),
		field.String("display_name").
			Comment("provider supplied display name").
			Optional().
			Annotations(
				entgql.OrderField("display_name"),
			),
		field.String("given_name").
			Comment("first name reported by the provider").
			Optional().
			Nillable(),
		field.String("family_name").
			Comment("last name reported by the provider").
			Optional().
			Nillable(),
		field.String("job_title").
			Comment("title captured at sync time").
			Optional().
			Nillable(),
		field.String("department").
			Comment("department captured at sync time").
			Optional().
			Nillable(),
		field.String("organization_unit").
			Comment("organizational unit or OU path the account lives under").
			Optional().
			Nillable(),
		field.Enum("account_type").
			Comment("type of principal represented in the directory").
			GoType(enums.DirectoryAccountType("")).
			Default(enums.DirectoryAccountTypeUser.String()).
			Optional(),
		field.Enum("status").
			Comment("lifecycle status returned by the directory").
			GoType(enums.DirectoryAccountStatus("")).
			Default(enums.DirectoryAccountStatusActive.String()),
		field.Enum("mfa_state").
			Comment("multi-factor authentication state reported by the directory").
			GoType(enums.DirectoryAccountMFAState("")).
			Default(enums.DirectoryAccountMFAStateUnknown.String()),
		field.String("last_seen_ip").
			Comment("last IP address observed by the provider, if any").
			Optional().
			Nillable(),
		field.Time("last_login_at").
			Comment("timestamp of the most recent login reported by the provider").
			Optional().
			Nillable(),
		field.Time("observed_at").
			Comment("time when this snapshot was recorded").
			Default(time.Now).
			Immutable(),
		field.String("profile_hash").
			Comment("hash of the normalized profile payload for change detection").
			Default(""),
		field.JSON("profile", map[string]any{}).
			Comment("flattened attribute bag used for filtering/diffing").
			Optional(),
		field.String("raw_profile_file_id").
			Comment("object storage file identifier that holds the raw upstream payload").
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

// Mixin of the DirectoryAccount
func (d DirectoryAccount) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:            "DAC",
		excludeSoftDelete: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(d),
		},
	}.getMixins(d)
}

// Edges of the DirectoryAccount
func (d DirectoryAccount) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Integration{},
			field:      "integration_id",
			required:   true,
			immutable:  true,
			comment:    "integration that owns this directory account",
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			edgeSchema: DirectorySyncRun{},
			field:      "directory_sync_run_id",
			required:   true,
			immutable:  true,
			comment:    "sync run that produced this snapshot",
		}),
		edge.To("groups", DirectoryGroup.Type).
			Annotations(
				entgql.RelayConnection(),
				entgql.QueryField(),
				entgql.MultiOrder(),
				accessmap.EdgeNoAuthCheck(),
			).
			Through("memberships", DirectoryMembership.Type),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: DirectoryAccount{},
			edgeSchema: WorkflowObjectRef{},
			name:       "workflow_object_refs",
			ref:        "directory_account",
		}),
	}
}

// Indexes of the DirectoryAccount
func (DirectoryAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("integration_id", "external_id", "directory_sync_run_id").
			Unique(),
		index.Fields("directory_sync_run_id", "canonical_email"),
		index.Fields("integration_id", "canonical_email"),
		index.Fields(ownerFieldName, "canonical_email"),
	}
}

// Policy of the DirectoryAccount
func (d DirectoryAccount) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the DirectoryAccount
func (d DirectoryAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.QueryField(),
		entx.CascadeThroughAnnotationField(
			[]entx.ThroughCleanup{
				{
					Field:   "DirectoryAccount",
					Through: "DirectoryMembership",
				},
			},
		),
	}
}
