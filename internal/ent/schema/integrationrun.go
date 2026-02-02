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

	"github.com/theopenlane/core/common/models"
)

// IntegrationRun holds the schema definition for integration execution history.
type IntegrationRun struct {
	SchemaFuncs

	ent.Schema
}

// SchemaIntegrationRun is the name of the IntegrationRun schema.
const SchemaIntegrationRun = "integration_run"

// Name returns the name of the IntegrationRun schema.
func (IntegrationRun) Name() string {
	return SchemaIntegrationRun
}

// GetType returns the type of the IntegrationRun schema.
func (IntegrationRun) GetType() any {
	return IntegrationRun.Type
}

// PluralName returns the plural name of the IntegrationRun schema.
func (IntegrationRun) PluralName() string {
	return pluralize.NewClient().Plural(SchemaIntegrationRun)
}

// Fields of the IntegrationRun.
func (IntegrationRun) Fields() []ent.Field {
	return []ent.Field{
		field.String("integration_id").
			Comment("integration connection this run belongs to").
			Optional(),
		field.String("operation_name").
			Comment("operation identifier executed for this run").
			Optional().
			Annotations(
				entgql.OrderField("OPERATION_NAME"),
			),
		field.String("operation_kind").
			Comment("operation category executed for this run").
			Optional().
			Annotations(
				entgql.OrderField("OPERATION_KIND"),
			),
		field.String("run_type").
			Comment("run type such as RUN or SYNC").
			Optional().
			Annotations(
				entgql.OrderField("RUN_TYPE"),
			),
		field.String("status").
			Comment("status of the run").
			Optional().
			Annotations(
				entgql.OrderField("STATUS"),
			),
		field.Time("started_at").
			Comment("when the run started").
			Default(time.Now).
			Annotations(
				entgql.OrderField("STARTED_AT"),
			),
		field.Time("finished_at").
			Comment("when the run completed").
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("FINISHED_AT"),
			),
		field.Int("duration_ms").
			Comment("run duration in milliseconds").
			Optional().
			Annotations(
				entgql.OrderField("DURATION_MS"),
			),
		field.String("request_file_id").
			Comment("file reference for the run request payload").
			Optional(),
		field.String("response_file_id").
			Comment("file reference for the run response payload").
			Optional(),
		field.String("event_id").
			Comment("event reference for this run").
			Optional(),
		field.String("summary").
			Comment("summary of the run outcome").
			Optional(),
		field.Text("error").
			Comment("error details for failed runs").
			Optional(),
		field.JSON("metrics", map[string]any{}).
			Comment("structured metrics and outputs for the run").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
	}
}

// Indexes of the IntegrationRun.
func (IntegrationRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("integration_id", "started_at").
			Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Edges of the IntegrationRun.
func (r IntegrationRun) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Integration{},
			field:      "integration_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Integration{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: r,
			field:      "request_file_id",
			name:       "request_file",
			t:          File.Type,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: r,
			field:      "response_file_id",
			name:       "response_file",
			t:          File.Type,
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: r,
			edgeSchema: Event{},
			field:      "event_id",
		}),
	}
}

// Mixin of the IntegrationRun.
func (r IntegrationRun) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeTags: true,
		excludeAnnotations: true,
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[IntegrationRun](r,
				withOrganizationOwnerServiceOnly(true),
			),
		},
	}.getMixins(r)
}

// Modules of the IntegrationRun.
func (IntegrationRun) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogBaseModule,
	}
}

// Annotations of the IntegrationRun.
func (r IntegrationRun) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entgql.Skip(entgql.SkipAll),
		entx.SchemaGenSkip(true),
		entx.QueryGenSkip(true),
		history.Annotations{
			Exclude: true,
		},
	}
}

// Policy of the IntegrationRun.
//func (IntegrationRun) Policy() ent.Policy {
//	return policy.NewPolicy(
//		policy.WithMutationRules(
//			rule.AllowMutationIfSystemAdmin(),
//		),
//	)
//}
