package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/entx/accessmap"
)

// Scan holds the schema definition for scan records used for domain, vulnerability, or provider scans.
type Scan struct {
	SchemaFuncs
	ent.Schema
}

// SchemaScan is the name of the Scan schema
const SchemaScan = "scan"

// Name returns the name of the Scan schema
func (Scan) Name() string {
	return SchemaScan
}

// GetType returns the type of the Scan schema
func (Scan) GetType() any {
	return Scan.Type
}

// PluralName returns the plural name of the Scan schema
func (Scan) PluralName() string {
	return pluralize.NewClient().Plural(SchemaScan)
}

// Fields of the Scan
func (Scan) Fields() []ent.Field {
	return []ent.Field{
		field.String("target").
			Comment("the target of the scan, e.g., a domain name or IP address, codebase").
			Annotations(entx.FieldSearchable()).
			NotEmpty(),
		field.Enum("scan_type").
			Comment("the type of scan, e.g., domain scan, vulnerability scan, provider scan").
			Annotations(entgql.OrderField("SCAN_TYPE"), entx.FieldSearchable()).
			GoType(enums.ScanType("")).
			Default(enums.ScanTypeDomain.String()),
		field.JSON("metadata", map[string]any{}).
			Comment("additional metadata for the scan, e.g., scan configuration, options, etc").
			Optional(),
		field.Time("scan_date").
			Comment("when the scan was executed").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("scan_date"),
			),
		field.String("scan_schedule").
			Comment("cron schedule that governs the scan cadence, in cron 6-field syntax").
			GoType(models.Cron("")).
			Optional().
			Nillable().
			Annotations(
				entgql.Skip(entgql.SkipWhereInput | entgql.SkipOrderField),
			).
			Validate(func(s string) error {
				if s == "" {
					return nil
				}

				c := models.Cron(s)

				return c.Validate()
			}),
		field.Time("next_scan_run_at").
			Comment("when the scan is scheduled to run next").
			GoType(models.DateTime{}).
			Optional().
			Nillable().
			Annotations(
				entgql.OrderField("next_scan_run_at"),
			),
		field.String("performed_by").
			Comment("who performed the scan when no user or group is linked").
			Optional(),
		field.String("performed_by_user_id").
			Comment("the user id that performed the scan").
			Optional(),
		field.String("performed_by_group_id").
			Comment("the group id that performed the scan").
			Optional(),
		field.String("generated_by_platform_id").
			Comment("the platform that generated the scan").
			Optional(),
		field.Strings("vulnerability_ids").
			Comment("identifiers of vulnerabilities discovered during the scan").
			Default([]string{}).
			Optional(),
		field.Enum("status").
			Comment("the status of the scan, e.g., processing, completed, failed").
			GoType(enums.ScanStatus("")).
			Default(enums.ScanStatusProcessing.String()).
			Annotations(entgql.OrderField("STATUS"), entx.FieldSearchable()),
	}
}

// Mixin of the Scan
func (s Scan) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s),
			newGroupPermissionsMixin(),
			newResponsibilityMixin(s, withReviewedBy(), withAssignedTo()),
			newCustomEnumMixin(s, withEnumFieldName("environment"), withGlobalEnum()),
			newCustomEnumMixin(s, withEnumFieldName("scope"), withGlobalEnum()),
		},
	}.getMixins(s)
}

// Edges of the Scan
func (s Scan) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(s, Asset{}),
		defaultEdgeToWithPagination(s, Entity{}),
		defaultEdgeToWithPagination(s, Evidence{}),
		defaultEdgeToWithPagination(s, File{}),
		defaultEdgeToWithPagination(s, Remediation{}),
		defaultEdgeToWithPagination(s, ActionPlan{}),
		defaultEdgeToWithPagination(s, Task{}),
		defaultEdgeFromWithPagination(s, Platform{}),
		defaultEdgeFromWithPagination(s, Vulnerability{}),
		defaultEdgeFromWithPagination(s, Control{}),
		uniqueEdgeFrom(&edgeDefinition{
			fromSchema: s,
			name:       "generated_by_platform",
			t:          Platform.Type,
			field:      "generated_by_platform_id",
			ref:        "generated_scans",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Platform{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: s,
			name:       "performed_by_user",
			t:          User.Type,
			field:      "performed_by_user_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(User{}.Name()),
			},
		}),
		uniqueEdgeTo(&edgeDefinition{
			fromSchema: s,
			name:       "performed_by_group",
			t:          Group.Type,
			field:      "performed_by_group_id",
			annotations: []schema.Annotation{
				accessmap.EdgeViewCheck(Group{}.Name()),
			},
		}),
	}
}

// Policy of the Scan
func (s Scan) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithMutationRules(
			rule.AllowMutationIfSystemAdmin(),
			policy.CheckOrgWriteAccess(),
		),
	)
}

// Modules this schema has access to
func (Scan) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogVulnerabilityManagementModule,
		models.CatalogComplianceModule,
	}
}

// Annotations of the Scan
func (Scan) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}
