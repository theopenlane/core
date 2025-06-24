package schema

import (
	"entgo.io/contrib/entgql"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/gertd/go-pluralize"

	"github.com/theopenlane/core/internal/ent/privacy/policy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/enums"
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
		field.String("target").NotEmpty(),
		field.Enum("scan_type").
			GoType(enums.ScanType("")).
			Default(enums.ScanTypeDomain.String()),
		field.JSON("metadata", map[string]any{}).
			Optional(),
		field.Enum("status").
			GoType(enums.ScanStatus("")).
			Default(enums.ScanStatusProcessing.String()).
			Annotations(entgql.OrderField("STATUS")),
	}
}

func (s Scan) Mixin() []ent.Mixin {
	return mixinConfig{
		additionalMixins: []ent.Mixin{
			newOrgOwnedMixin(s),
			newGroupPermissionsMixin(),
		},
	}.getMixins()
}

func (s Scan) Edges() []ent.Edge {
	return []ent.Edge{
		defaultEdgeToWithPagination(s, Asset{}),
		defaultEdgeToWithPagination(s, Risk{}),
	}
}

func (Scan) Policy() ent.Policy {
	return policy.NewPolicy(
		policy.WithQueryRules(
			policy.CheckOrgReadAccess(),
			rule.AllowQueryIfSystemAdmin(),
		),
		policy.WithMutationRules(
			policy.CheckOrgWriteAccess(),
			rule.AllowMutationIfSystemAdmin(),
		),
	)
}
