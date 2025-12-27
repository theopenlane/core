package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gertd/go-pluralize"
	"github.com/theopenlane/entx/accessmap"
	"github.com/theopenlane/iam/entfga"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
)

// FindingControl defines the mapping between findings and controls.
type FindingControl struct {
	SchemaFuncs

	ent.Schema
}

// SchemaFindingControl is the name of the finding control schema.
const SchemaFindingControl = "finding_control"

// Name returns the name of the finding control schema.
func (FindingControl) Name() string {
	return SchemaFindingControl
}

// GetType returns the type of the finding control schema.
func (FindingControl) GetType() any {
	return FindingControl.Type
}

// PluralName returns the plural name of the finding control schema.
func (FindingControl) PluralName() string {
	return pluralize.NewClient().Plural(SchemaFindingControl)
}

// Fields returns finding control fields.
func (FindingControl) Fields() []ent.Field {
	return []ent.Field{
		field.String("finding_id").
			Immutable().
			Comment("the id of the finding associated with the control"),
		field.String("control_id").
			Immutable().
			Comment("the id of the control mapped to the finding when it exists in the catalog"),
		field.String("standard_id").
			Optional().
			Immutable().
			Comment("the id of the standard that the control belongs to when it exists in the catalog"),
		field.String("external_standard").
			Comment("external identifier for the standard provided by the source system such as iso or hipaa").
			Optional(),
		field.String("external_standard_version").
			Comment("version for the external standard provided by the source system").
			Optional(),
		field.String("external_control_id").
			Comment("control identifier provided by the source system such as A.5.10").
			Optional(),
		field.String("source").
			Comment("the integration source that provided the mapping").
			Optional(),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata about the control mapping from the source system").
			Optional(),
		field.Time("discovered_at").
			GoType(models.DateTime{}).
			Comment("timestamp when the mapping was first observed").
			Optional().
			Nillable(),
	}
}

// Edges of the FindingControl
func (fc FindingControl) Edges() []ent.Edge {
	return []ent.Edge{
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: fc,
			edgeSchema: Finding{},
			field:      "finding_id",
			required:   true,
			immutable:  true,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: fc,
			edgeSchema: Control{},
			field:      "control_id",
			required:   true,
			immutable:  true,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: fc,
			edgeSchema: Standard{},
			field:      "standard_id",
			immutable:  true,
			annotations: []schema.Annotation{
				accessmap.EdgeNoAuthCheck(),
			},
		}),
	}
}

// Indexes of the FindingControl
func (FindingControl) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("finding_id", "control_id").
			Unique().
			Annotations(),
	}
}

// Mixin of the FindingControl
func (FindingControl) Mixin() []ent.Mixin {
	return mixinConfig{
		excludeSoftDelete: true,
		excludeTags:       true,
	}.getMixins(FindingControl{})
}

// Modules of the FindingControl
func (FindingControl) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogVulnerabilityManagementModule,
		models.CatalogComplianceModule,
	}
}

// Annotations of the FindingControl
func (FindingControl) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entfga.SelfAccessChecks(),
	}
}

// Policy of the FindingControl
func (FindingControl) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
		),
	)
}
