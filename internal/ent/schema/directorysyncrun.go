package schema

import (
	"time"

	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/history"
)

// DirectorySyncRun captures the execution metadata for a directory ingestion job
type DirectorySyncRun struct {
	SchemaFuncs

	ent.Schema
}

// SchemaDirectorySyncRun is the canonical schema name
const SchemaDirectorySyncRun = "directory_sync_run"

// Name returns the schema name
func (DirectorySyncRun) Name() string {
	return SchemaDirectorySyncRun
}

// GetType returns the ent type
func (DirectorySyncRun) GetType() any {
	return DirectorySyncRun.Type
}

// PluralName returns the pluralized name
func (DirectorySyncRun) PluralName() string {
	return pluralize.NewClient().Plural(SchemaDirectorySyncRun)
}

// Fields defines the DirectorySyncRun attributes
func (DirectorySyncRun) Fields() []ent.Field {
	return []ent.Field{
		field.String("integration_id").
			Comment("integration this sync run executed for").
			NotEmpty().
			Immutable(),
		field.String("platform_id").
			Comment("optional platform associated with this sync run").
			Optional().
			NotEmpty().
			Immutable(),
		field.Enum("status").
			Comment("current state of the sync run").
			GoType(enums.DirectorySyncRunStatus("")).
			Default(enums.DirectorySyncRunStatusPending.String()),
		field.Time("started_at").
			Comment("time the sync started").
			Default(time.Now).
			Annotations(
				entgql.OrderField("started_at"),
			),
		field.Time("completed_at").
			Comment("time the sync finished").
			Optional().
			Nillable(),
		field.String("source_cursor").
			Comment("cursor or checkpoint returned by the provider for the next run").
			Optional().
			Nillable(),
		field.Int("full_count").
			Comment("total records processed during this run").
			Default(0),
		field.Int("delta_count").
			Comment("number of records that changed compared to the prior run").
			Default(0),
		field.Text("error").
			Comment("serialized error information when the run failed").
			Optional().
			Nillable(),
		field.String("raw_manifest_file_id").
			Comment("object storage file identifier for the manifest captured during the run").
			Optional().
			Nillable(),
		field.JSON("stats", map[string]any{}).
			Comment("additional provider-specific stats for the run").
			Optional(),
	}
}

// Mixin of the DirectorySyncRun
func (r DirectorySyncRun) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix:            "DSR",
		excludeTags:       true,
		excludeSoftDelete: true,
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(r),
			newCustomEnumMixin(r, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(r, withEnumFieldName("scope"), withGlobalEnum()),
		},
	}.getMixins(r)
}

// Edges of the DirectorySyncRun
func (r DirectorySyncRun) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Integration{},
			field:      "integration_id",
			required:   true,
			immutable:  true,
			comment:    "integration that executed this sync run",
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Platform{},
			field:      "platform_id",
			immutable:  true,
			comment:    "platform associated with this sync run",
		}),
		defaultEdgeToWithPagination(r, DirectoryAccount{}),
		defaultEdgeToWithPagination(r, DirectoryGroup{}),
		defaultEdgeToWithPagination(r, DirectoryMembership{}),
	}
}

// Indexes of the DirectorySyncRun
func (DirectorySyncRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("integration_id", "started_at").
			Annotations(),
		index.Fields("platform_id", "started_at"),
	}
}

// Policy constrains access to DirectorySyncRun
func (r DirectorySyncRun) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Annotations of the DirectorySyncRun
func (DirectorySyncRun) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entx.SchemaSearchable(false),
		history.Annotations{
			Exclude: true,
		},
	}
}
