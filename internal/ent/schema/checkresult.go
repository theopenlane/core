package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/entfga"
)

// CheckResult holds the schema definition for the CheckResult entity
type CheckResult struct {
	SchemaFuncs

	ent.Schema
}

// SchemaCheckResult is the name of the schema in snake case
const SchemaCheckResult = "check_result"

// Name is the name of the schema in snake case
func (CheckResult) Name() string {
	return SchemaCheckResult
}

// GetType returns the type of the schema
func (CheckResult) GetType() any {
	return CheckResult.Type
}

// PluralName returns the plural name of the schema
func (CheckResult) PluralName() string {
	return pluralize.NewClient().Plural(SchemaCheckResult)
}

// Fields of the CheckResult
func (CheckResult) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			Comment("current status of the control").
			Annotations(
				entgql.OrderField("STATUS"),
			).
			GoType(enums.CheckStatus("")).
			Default(enums.CheckStatusUnknown.String()),
		field.String("source").
			Annotations(
				entgql.OrderField("source"),
			).
			Comment("source that set the check result"),
		field.Time("last_observed_at").
			Comment("timestamp the result was last updated").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("observed_at"),
			),
		field.String("external_uri").
			Comment("link to the result in the source system").
			Optional().
			Annotations(
				entx.IntegrationMappingField(),
			),
		field.Text("details").Comment("optional details of the result").Optional().Nillable(),
		field.String("parent_external_id").
			Comment("external parent reference id for the aggregate rule, e.g. in aws config this is the config rule name").
			Optional().
			Annotations(
				entx.IntegrationMappingField().UpsertKey().LookupKey(),
			),
		field.String("integration_id").
			Comment("integration that owns this directory group").
			Optional().
			Immutable().
			Annotations(
				entx.IntegrationMappingField().UpsertKey().FromIntegration(),
			),
	}
}

// Mixin of the CheckResult
func (c CheckResult) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.CheckResult](c,
				withParents(
					Control{},
				),
			),
			newGroupPermissionsMixin(),
		},
	}.getMixins(c)
}

// Edges of the CheckResult
func (c CheckResult) Edges() []ent.Edge {
	return []ent.Edge{
		edgeToWithPagination(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Control{},
			annotations: []schema.Annotation{
				entx.CSVRef().FromColumn("ControlRefCodes").MatchOn("ref_code"),
			},
		}),
		defaultEdgeFromWithPagination(c, Finding{}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: c,
			edgeSchema: Integration{},
			field:      "integration_id",
			required:   false,
			immutable:  true,
			comment:    "integration that owns this control health",
		}),
	}
}

// Indexes of the CheckResult
func (CheckResult) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the CheckResult
func (CheckResult) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.NewExportable(
			entx.WithOrgOwned(),
			entx.WithSystemOwned(),
		),
		entx.IntegrationMappingSchema().StockPersist(),
	}
}

// Hooks of the CheckResult
func (CheckResult) Hooks() []ent.Hook {
	return []ent.Hook{}
}

// Interceptors of the CheckResult
func (CheckResult) Interceptors() []ent.Interceptor {
	return []ent.Interceptor{}
}

// Modules this schema has access to
func (CheckResult) Modules() []models.OrgModule {
	return []models.OrgModule{}
}

// Policy of the CheckResult
func (CheckResult) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			policy.CheckOrgWriteAccess(),
			policy.CanCreateObjectsUnderParents([]string{
				Control{}.PluralName(),
			}),
			entfga.CheckEditAccess[*generated.CheckResultMutation](),
		),
	)
}
