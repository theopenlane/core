package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/mixin"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// Finding defines the finding schema.
type Finding struct {
	SchemaFuncs

	ent.Schema
}

// SchemaFinding is the name of the finding schema.
const SchemaFinding = "finding"

// Name returns the name of the finding schema.
func (Finding) Name() string {
	return SchemaFinding
}

// GetType returns the type of the finding schema.
func (Finding) GetType() any {
	return Finding.Type
}

// PluralName returns the plural name of the finding schema.
func (Finding) PluralName() string {
	return pluralize.NewClient().Plural(SchemaFinding)
}

// Fields returns finding fields.
func (Finding) Fields() []ent.Field {
	return []ent.Field{
		field.String("external_id").
			Comment("external identifier from the integration source for the finding").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("external_id"),
			),
		field.String("external_owner_id").
			Comment("the owner of the finding").
			Optional().
			Annotations(
				entx.FieldSearchable(),
				entgql.OrderField("external_owner_id"),
			),
		field.String("source").
			Comment("system that produced the finding, e.g. gcpscc").
			Optional(),
		field.String("resource_name").
			Comment("resource identifier provided by the source system").
			Optional(),
		field.String("display_name").
			Comment("display name for the finding when provided by the source").
			Optional(),
		field.String("state").
			Comment("state reported by the source system, such as ACTIVE or INACTIVE").
			Optional(),
		field.String("category").
			Comment("primary category of the finding").
			Optional().
			Annotations(
				entgql.OrderField("category"),
			),
		field.Strings("categories").
			Comment("normalized categories for the finding").
			Optional().
			Default([]string{}),
		field.String("finding_class").
			Comment("classification provided by the source, e.g. MISCONFIGURATION").
			Optional(),
		field.String("severity").
			Comment("severity label for the finding").
			Optional().
			Annotations(
				entgql.OrderField("severity"),
				entx.FieldSearchable(),
			),
		field.Float("numeric_severity").
			Comment("numeric severity score for the finding if provided").
			Optional(),
		field.Float("score").
			Comment("aggregated score such as CVSS for the finding").
			Optional(),
		field.Float("impact").
			Comment("impact score or rating for the finding").
			Optional(),
		field.Float("exploitability").
			Comment("exploitability score or rating for the finding").
			Optional(),
		field.String("priority").
			Comment("priority assigned to the finding").
			Optional(),
		field.Bool("open").
			Comment("indicates if the finding is still open").
			Default(true).
			Optional(),
		field.Bool("blocks_production").
			Comment("true when the finding blocks production changes").
			Optional(),
		field.Bool("production").
			Comment("true when the finding affects production systems").
			Optional(),
		field.Bool("public").
			Comment("true when the finding is publicly disclosed").
			Optional(),
		field.Bool("validated").
			Comment("true when the finding has been validated by the security team").
			Optional(),
		field.String("assessment_id").
			Comment("identifier for the assessment that generated the finding").
			Optional(),
		field.Text("description").
			Comment("long form description of the finding").
			Optional(),
		field.Text("recommendation").
			Comment("short recommendation text from the source system (deprecated upstream)").
			Optional(),
		field.Text("recommended_actions").
			Comment("markdown formatted remediation guidance for the finding").
			Optional(),
		field.Strings("references").
			Comment("reference links for the finding").
			Optional().
			Default([]string{}),
		field.Strings("steps_to_reproduce").
			Comment("steps required to reproduce the finding").
			Optional().
			Default([]string{}),
		field.Strings("targets").
			Comment("targets impacted by the finding such as projects or applications").
			Optional().
			Default([]string{}),
		field.JSON("target_details", map[string]any{}).
			Comment("structured details about the impacted targets").
			Optional(),
		field.String("vector").
			Comment("attack vector string such as a CVSS vector").
			Optional(),
		field.Int("remediation_sla").
			Comment("remediation service level agreement in days").
			Optional(),
		field.String("status").
			Comment("lifecycle status of the finding").
			Optional(),
		field.Time("event_time").
			Comment("timestamp when the finding was last observed by the source").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.Time("reported_at").
			Comment("timestamp when the finding was first reported by the source").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.Time("source_updated_at").
			Comment("timestamp when the source last updated the finding").
			GoType(models.DateTime{}).
			Optional().
			Nillable(),
		field.String("external_uri").
			Comment("link to the finding in the source system").
			Optional(),
		field.JSON("metadata", map[string]any{}).
			Comment("raw metadata payload for the finding from the source system").
			Optional(),
		field.JSON("raw_payload", map[string]any{}).
			Comment("raw payload received from the integration for auditing and troubleshooting").
			Optional(),
	}
}

// Edges of the Finding
func (f Finding) Edges() []ent.Edge {
	return []ent.Edge{
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: f,
			edgeSchema: Integration{},
			comment:    "integration that produced the finding",
		}),
		defaultEdgeToWithPagination(f, Vulnerability{}),
		defaultEdgeToWithPagination(f, ActionPlan{}),
		edge.To("controls", Control.Type).
			Annotations(
				entgql.RelayConnection(),
				entgql.QueryField(),
				entgql.MultiOrder(),
			).
			Through("control_mappings", FindingControl.Type),
		defaultEdgeToWithPagination(f, Subcontrol{}),
		defaultEdgeToWithPagination(f, Risk{}),
		defaultEdgeToWithPagination(f, Program{}),
		defaultEdgeToWithPagination(f, Asset{}),
		defaultEdgeToWithPagination(f, Entity{}),
		defaultEdgeToWithPagination(f, Scan{}),
		defaultEdgeToWithPagination(f, Task{}),
		defaultEdgeToWithPagination(f, DirectoryAccount{}),
		defaultEdgeToWithPagination(f, IdentityHolder{}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: f,
			edgeSchema: Remediation{},
			comment:    "remediation efforts tracked against the finding",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: f,
			edgeSchema: Review{},
			comment:    "reviews performed for this finding",
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: f,
			name:       "comments",
			t:          Note.Type,
			comment:    "discussion threads associated with the finding",
			annotations: []schema.Annotation{
				accessmap.EdgeAuthCheck(Note{}.Name()),
			},
		}),
		edgeToWithPagination(&edgeDefinition{
			fromSchema: f,
			edgeSchema: File{},
			comment:    "supporting files or evidence for the finding",
		}),
		edgeFromWithPagination(&edgeDefinition{
			fromSchema: f,
			edgeSchema: WorkflowObjectRef{},
			name:       "workflow_object_refs",
			ref:        "finding",
		}),
	}
}

// Mixin of the Finding
func (f Finding) Mixin() []ent.Mixin {
	return mixinConfig{
		prefix: "FIND",
		additionalMixins: []ent.Mixin{
			newObjectOwnedMixin[generated.Finding](f,
				withParents(
					Program{},
					Control{},
					Subcontrol{},
					Risk{},
					Asset{},
					Entity{},
					Scan{},
					DirectoryAccount{},
					IdentityHolder{},
				),
				withOrganizationOwner(true),
			),
			newGroupPermissionsMixin(),
			mixin.NewSystemOwnedMixin(mixin.SkipTupleCreation()),
			newCustomEnumMixin(f, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(f, withEnumFieldName("scope"), withGlobalEnum()),
		},
	}.getMixins(f)
}

// Indexes of the Finding
func (Finding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("external_id", "external_owner_id", ownerFieldName).
			Unique().
			Annotations(
				entsql.IndexWhere("deleted_at is NULL"),
			),
	}
}

// Annotations of the Finding
func (Finding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
		entx.NewExportable(
			entx.WithOrgOwned(),
			entx.WithSystemOwned(),
		),
	}
}

// Policy of the Finding
func (f Finding) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckCreateAccess(),
			entfga.CheckEditAccess[*generated.FindingMutation](),
		),
	)
}

func (Finding) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogVulnerabilityManagementModule,
		models.CatalogComplianceModule,
	}
}
