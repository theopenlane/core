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

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/hooks"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/validator"
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
			Comment("optional integration that owns this directory account when sourced by an integration").
			Optional().
			NotEmpty().
			Immutable(),
		field.String("directory_sync_run_id").
			Comment("optional sync run that produced this snapshot").
			Optional().
			NotEmpty().
			Immutable(),
		field.String("platform_id").
			Comment("optional platform associated with this directory account").
			Optional().
			NotEmpty().
			Immutable(),
		field.String("identity_holder_id").
			Comment("deduplicated identity holder linked to this directory account").
			Optional().
			Nillable().
			Annotations(
				entx.CSVRef().FromColumn("DirectoryAccountIdentityHolderEmail").MatchOn("email"),
			),
		field.String("directory_name").
			Comment("directory source label set by the integration (e.g. googleworkspace, github, slack)").
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("directory_name"),
			),
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
		field.String("avatar_remote_url").
			Comment("URL of the avatar supplied by the directory provider").
			MaxLen(2048). //nolint:mnd
			Validate(validator.ValidateURL()).
			Optional().
			Nillable(),
		field.String("avatar_local_file_id").
			Comment("local avatar file identifier, takes precedence over avatar_remote_url").
			Optional().
			Nillable().
			Annotations(
				entgql.Skip(entgql.SkipMutationCreateInput, entgql.SkipMutationUpdateInput),
			),
		field.Time("avatar_updated_at").
			Comment("time the directory account avatar was last updated").
			Default(time.Now).
			UpdateDefault(time.Now).
			Optional().
			Nillable(),
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
			newCustomEnumMixin(d, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(d, withEnumFieldName("scope"), withGlobalEnum()),
		},
	}.getMixins(d)
}

// Edges of the DirectoryAccount
func (d DirectoryAccount) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Integration{},
			field:      "integration_id",
			immutable:  true,
			comment:    "integration that owns this directory account",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: d,
			edgeSchema: DirectorySyncRun{},
			field:      "directory_sync_run_id",
			immutable:  true,
			comment:    "sync run that produced this snapshot",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: d,
			edgeSchema: Platform{},
			field:      "platform_id",
			immutable:  true,
			comment:    "platform associated with this directory account",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: d,
			edgeSchema: IdentityHolder{},
			field:      "identity_holder_id",
			comment:    "identity holder linked to this directory account",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(IdentityHolder{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: d,
			name:       "avatar_file",
			t:          File.Type,
			field:      "avatar_local_file_id",
			comment:    "local avatar file for the directory account",
		}),
		edge.To("groups", DirectoryGroup.Type).
			Annotations(
				entgql.RelayConnection(),
				entgql.QueryField(),
				entgql.MultiOrder(),
				accessmap.EdgeNoAuthCheck(),
			).
			Through("memberships", DirectoryMembership.Type),
		defaultEdgeFromWithPagination(d, Finding{}),
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
		index.Fields("platform_id", "external_id"),
		index.Fields("directory_sync_run_id", "canonical_email"),
		index.Fields("integration_id", "canonical_email"),
		index.Fields("platform_id", "canonical_email"),
		index.Fields("identity_holder_id"),
		index.Fields("identity_holder_id", "directory_name"),
		index.Fields(ownerFieldName, "canonical_email"),
	}
}

// Hooks of the DirectoryAccount
func (DirectoryAccount) Hooks() []ent.Hook {
	return []ent.Hook{
		hooks.HookDirectoryAccount(),
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
		entx.IntegrationMappingSchema().
			UpsertKeys("external_id", "canonical_email"),
	}
}
