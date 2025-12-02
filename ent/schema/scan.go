package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/shared/enums"
	"github.com/theopenlane/shared/models"
	"github.com/theopenlane/ent/privacy/policy"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/entx"
)

// Scan holds the schema definition for scan records used for domain, vulnerability, or provider scans.
type Scan struct {
	SchemaFuncs
	ent.Schema
}

const SchemaScan = "scan"

func (Scan) Name() string       { return SchemaScan }
func (Scan) GetType() any       { return Scan.Type }
func (Scan) PluralName() string { return pluralize.NewClient().Plural(SchemaScan) }

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
		field.Enum("status").
			Comment("the status of the scan, e.g., processing, completed, failed").
			GoType(enums.ScanStatus("")).
			Default(enums.ScanStatusProcessing.String()).
			Annotations(entgql.OrderField("STATUS"), entx.FieldSearchable()),
	}
}

func (s Scan) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s),
			newGroupPermissionsMixin(),
		},
	}.getMixins(s)
}

func (s Scan) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(s, Asset{}),
		defaultEdgeToWithPagination(s, Entity{}),
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

func (Scan) Modules() []models.OrgModule {
	return []models.OrgModule{
		models.CatalogVulnerabilityManagementModule,
	}
}

// Annotations of the Scan
func (Scan) Annotations() []schema.Annotation {
	return []schema.Annotation{}
}
