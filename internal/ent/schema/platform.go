package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
)

// Platform holds the schema definition for the Platform entity
type Platform struct {
	SchemaFuncs

	ent.Schema
}

// SchemaPlatform is the name of the Platform schema
const SchemaPlatform = "platform"

// Name returns the name of the Platform schema
func (Platform) Name() string {
	return SchemaPlatform
}

// GetType returns the type of the Platform schema
func (Platform) GetType() any {
	return Platform.Type
}

// PluralName returns the plural name of the Platform schema
func (Platform) PluralName() string {
	return pluralize.NewClient().Plural(SchemaPlatform)
}

// Fields of the Platform
func (Platform) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Comment("the name of the platform").
			NotEmpty().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("name"),
			),
		field.String("description").
			Comment("the description of the platform boundary").
			Optional(),
		field.String("business_purpose").
			Comment("the business purpose of the platform").
			Optional().
			Annotations(
				entgql.OrderField("business_purpose"),
				entx.FieldWorkflowEligible(),
			),
		field.Text("scope_statement").
			Comment("scope statement for the platform, used for narrative justification").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
				entx.FieldWorkflowEligible(),
			),
		field.Text("trust_boundary_description").
			Comment("description of the platform trust boundary").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Text("data_flow_summary").
			Comment("summary of platform data flows").
			Optional().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput),
			),
		field.Enum("status").
			Comment("the lifecycle status of the platform").
			GoType(enums.PlatformStatus("")).
			Default(enums.PlatformStatusActive.String()).
			Annotations(
				entgql.OrderField("STATUS"),
				entx.FieldWorkflowEligible(),
			),
		field.String("physical_location").
			Comment("physical location of the platform, if applicable").
			Optional().
			Annotations(
				entgql.OrderField("physical_location"),
			),
		field.String("region").
			Comment("the region where the platform operates or is hosted").
			Optional().
			Annotations(
				entgql.OrderField("region"),
			),
		field.Bool("contains_pii").
			Comment("whether the platform stores or processes PII").
			Default(false).
			Optional().
			Annotations(
				entgql.OrderField("contains_pii"),
				entx.FieldWorkflowEligible(),
			),
		field.Enum("source_type").
			Comment("the source of the platform record, e.g., manual, discovered, imported, api").
			GoType(enums.SourceType("")).
			Default(enums.SourceTypeManual.String()).
			Annotations(
				entgql.OrderField("SOURCE_TYPE"),
			),
		field.String("source_identifier").
			Comment("the identifier used by the source system for the platform").
			Optional().
			Annotations(
				entgql.OrderField("source_identifier"),
			),
		field.String("cost_center").
			Comment("cost center associated with the platform").
			Optional().
			Annotations(
				entgql.OrderField("cost_center"),
			),
		field.Float("estimated_monthly_cost").
			Comment("estimated monthly cost for the platform").
			Optional().
			Annotations(
				entgql.OrderField("estimated_monthly_cost"),
			),
		field.Time("purchase_date").
			Comment("purchase date for the platform").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("purchase_date"),
			),
		field.String("platform_owner_id").
			Comment("the id of the user who is responsible for this platform").
			Optional(),
		field.String("external_reference_id").
			Comment("external identifier for the platform from an upstream inventory").
			Optional().
			Annotations(
				entgql.OrderField("external_reference_id"),
			),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the platform").
			Optional(),
	}
}

// Mixin of the Platform
func (s Platform) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "PLT",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[Platform](s,
				withParents(Organization{}, Entity{}),
				withOrganizationOwner(true),
			),
			//			newGroupPermissionsMixin(),
			newResponsibilityMixin(
				s,
				withInternalOwner(),
				withBusinessOwner(),
				withTechnicalOwner(),
				withSecurityOwner(),
			),
			newCustomEnumMixin(s),
			newCustomEnumMixin(s, withEnumFieldName("data_classification")),
			newCustomEnumMixin(s, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(s, withEnumFieldName("scope"), withGlobalEnum()),
			newCustomEnumMixin(s, withEnumFieldName("access_model"), withGlobalEnum()),
			newCustomEnumMixin(s, withEnumFieldName("encryption_status"), withGlobalEnum()),
			newCustomEnumMixin(s, withEnumFieldName("security_tier"), withGlobalEnum()),
			newCustomEnumMixin(s, withEnumFieldName("criticality"), withGlobalEnum(), withWorkflowEnumEdges()),
			WorkflowApprovalMixin{},
		},
	}.getMixins(s)
}

// Edges of the Platform
func (s Platform) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(s, Asset{}),
		defaultEdgeToWithPagination(s, Entity{}),
		defaultEdgeToWithPagination(s, Evidence{}),
		defaultEdgeToWithPagination(s, File{}),
		defaultEdgeToWithPagination(s, Risk{}),
		defaultEdgeToWithPagination(s, Control{}),
		defaultEdgeToWithPagination(s, Assessment{}),
		defaultEdgeToWithPagination(s, Scan{}),
		defaultEdgeToWithPagination(s, Task{}),
		defaultEdgeToWithPagination(s, IdentityHolder{}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: s,
			edgeSchema: WorkflowObjectRef{},
			name:       "workflow_object_refs",
			ref:        "platform",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			name:       "source_assets",
			t:          Asset.Type,
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Asset{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			name:       "source_entities",
			t:          Entity.Type,
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Entity{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			name:       "out_of_scope_assets",
			t:          Asset.Type,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			name:       "out_of_scope_vendors",
			t:          Entity.Type,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			name:       "applicable_frameworks",
			t:          Standard.Type,
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: s,
			name:       "generated_scans",
			t:          Scan.Type,
		}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: s,
			name:       "platform_owner",
			t:          User.Type,
			field:      "platform_owner_id",
			ref:        "platforms_owned",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(User{}.Name()),
			},
		}),
	}
}

// Indexes of the Platform
func (Platform) Indexes() []ent.Index {
	return []ent.Index{
		// names should be unique, but ignore deleted names
		index.Fields("name", ownerFieldName).
			Unique().Annotations(entsql.IndexWhere("deleted_at is NULL")),
	}
}

// Modules this schema has access to
func (Platform) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogComplianceModule,
		models.CatalogEntityManagementModule,
	}
}

// Annotations of the Platform
//func (Platform) Annotations() []schema.Annotation {
//	return []schema.Annotation{
//		entfga.SelfAccessChecks(),
//	}
//}
//
//// Policy of the Platform
//func (Platform) Policy() ent.Policy {
//	return policy.NewPolicy(
//		policy.WithMutationRules(
//			policy.CheckCreateAccess(),
//			entfga.CheckEditAccess[*generated.PlatformMutation](),
//		),
//	)
//}
