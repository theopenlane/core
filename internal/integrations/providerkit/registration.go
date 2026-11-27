package providerkit

import (
	"context"
	"encoding/json"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/asset"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/v2/internal/ent/generated/risk"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/utils/jsonx"
)

const (
	// matchAllFilterExpr is the filter expression admitting every envelope to its mapping
	matchAllFilterExpr = "true"
	// externalIDField is the provenance mixin's external id column records link to each other through
	externalIDField = "external_id"
	// nameField is the name column records link to named targets such as groups and entity types through
	nameField = "name"
)

// edgeMatch is one target match-key field edges link through, with the payload key suffixes carrying its values
type edgeMatch struct {
	// field is the target match-key field
	field string
	// one follows a unique edge name in the payload key holding the one value it links to
	one string
	// many follows a to-many edge name in the payload key listing the values it links to
	many string
}

var (
	// externalIDMatch links edges through the target external id
	externalIDMatch = edgeMatch{field: externalIDField, one: "_external_id", many: "_external_ids"}
	// nameMatch links edges through the target name
	nameMatch = edgeMatch{field: nameField, one: "_name", many: "_names"}
)

// edgeMatches are the target match keys EdgeKeyLinks derives links through
var edgeMatches = []edgeMatch{externalIDMatch, nameMatch}

// DirectorySync is the stored config of a directory sync operation that can skip groups and memberships; a directory without group sync declares its own DirectorySync embedding only types.OperationSettings
type DirectorySync struct {
	types.OperationSettings
	// DisableGroupSync skips collecting groups and memberships, syncing only users
	DisableGroupSync bool `json:"disableGroupSync,omitempty" jsonschema:"title=Disable Group Sync,description=Only sync users and skip group and membership sync"`
}

// UpgradeFromSection returns an upgrade decoding T from the object stored under key when present, else from the stored document; it reads main's client config section for an operation whose main key differs from its camelCase name
//
// TODO: remove with the integration config column once every installation has been upgraded off main's client config
func UpgradeFromSection[T any](key string) func(context.Context, types.InstallationRequest, string, json.RawMessage) (T, error) {
	return func(_ context.Context, _ types.InstallationRequest, _ string, stored json.RawMessage) (T, error) {
		if section, ok := jsonx.DecodeObjectKey[T](stored, key); ok {
			return section, nil
		}

		return jsonx.Decode[T](stored)
	}
}

// DirectoryIngestContracts returns ingest contracts for directory accounts, groups, and memberships
func DirectoryIngestContracts() []types.IngestContract {
	return []types.IngestContract{
		{Schema: entityops.SchemaDirectoryAccount.Name},
		{Schema: entityops.SchemaDirectoryGroup.Name},
		{Schema: entityops.SchemaDirectoryMembership.Name},
	}
}

// DirectoryMappings returns account, group, and membership mappings linked by external id
func DirectoryMappings(accountExpr, groupExpr, membershipExpr string) []types.MappingRegistration {
	return []types.MappingRegistration{
		{
			Schema: entityops.SchemaDirectoryAccount.Name,
			Spec: types.MappingOverride{
				FilterExpr: matchAllFilterExpr,
				MapExpr:    accountExpr,
			},
		},
		{
			Schema: entityops.SchemaDirectoryGroup.Name,
			Spec: types.MappingOverride{
				FilterExpr: matchAllFilterExpr,
				MapExpr:    groupExpr,
			},
		},
		{
			Schema: entityops.SchemaDirectoryMembership.Name,
			Spec: types.MappingOverride{
				FilterExpr: matchAllFilterExpr,
				MapExpr:    membershipExpr,
				Links: []types.LinkRule{
					{
						TargetSchema: entityops.SchemaDirectoryAccount.Name,
						TargetField:  directoryaccount.FieldExternalID,
						SourceField:  entityops.DirectoryMembershipFields.DirectoryAccountID.InputKey,
					},
					{
						TargetSchema: entityops.SchemaDirectoryGroup.Name,
						TargetField:  directorygroup.FieldExternalID,
						SourceField:  entityops.DirectoryMembershipFields.DirectoryGroupID.InputKey,
					},
				},
			},
		},
	}
}

// FindingMapping returns the finding mapping linking findings to controls by ref code, assets by name and external id, and the internal owner identity holder by email
func FindingMapping(mapExpr string) types.MappingRegistration {
	return types.MappingRegistration{
		Schema: entityops.SchemaFinding.Name,
		Spec: types.MappingOverride{
			FilterExpr: matchAllFilterExpr,
			MapExpr:    mapExpr,
			Links: append([]types.LinkRule{
				{
					TargetSchema: entityops.SchemaControl.Name,
					TargetField:  control.FieldRefCode,
					SourceField:  entityops.FindingFields.Category.InputKey,
					SourceList:   entityops.FindingFields.Categories.InputKey,
				},
				{
					TargetSchema: entityops.SchemaAsset.Name,
					TargetField:  asset.FieldName,
					SourceField:  entityops.FindingFields.ResourceName.InputKey,
					SourceList:   entityops.FindingFields.Targets.InputKey,
				},
				internalOwnerIdentityHolderLink(finding.EdgeInternalOwnerIdentityHolder, entityops.FindingFields.InternalOwner.InputKey),
			}, externalIDMatch.links(entityops.SchemaFinding, []string{finding.EdgeAssets})...),
		},
	}
}

// CheckResultMapping returns the check result mapping linking check results to controls whose ref code is in the tags
func CheckResultMapping(mapExpr string) types.MappingRegistration {
	return types.MappingRegistration{
		Schema: entityops.SchemaCheckResult.Name,
		Spec: types.MappingOverride{
			FilterExpr: matchAllFilterExpr,
			MapExpr:    mapExpr,
			Links: []types.LinkRule{
				{
					TargetSchema: entityops.SchemaControl.Name,
					TargetField:  control.FieldRefCode,
					SourceList:   entityops.CheckResultFields.Tags.InputKey,
				},
			},
		},
	}
}

// VulnerabilityMapping returns the vulnerability mapping linking the internal owner to an identity holder by email and assets by edge keys
func VulnerabilityMapping(mapExpr string) types.MappingRegistration {
	return types.MappingRegistration{
		Schema: entityops.SchemaVulnerability.Name,
		Spec: types.MappingOverride{
			FilterExpr: matchAllFilterExpr,
			MapExpr:    mapExpr,
			Links: append([]types.LinkRule{
				internalOwnerIdentityHolderLink(vulnerability.EdgeInternalOwnerIdentityHolder, entityops.VulnerabilityFields.InternalOwner.InputKey),
			}, EdgeKeyLinks(entityops.SchemaVulnerability, vulnerability.EdgeAssets)...),
		},
	}
}

// RiskMapping returns the risk mapping linking assets by edge keys
func RiskMapping(mapExpr string) types.MappingRegistration {
	return types.MappingRegistration{
		Schema: entityops.SchemaRisk.Name,
		Spec: types.MappingOverride{
			FilterExpr: matchAllFilterExpr,
			MapExpr:    mapExpr,
			Links:      EdgeKeyLinks(entityops.SchemaRisk, risk.EdgeAssets),
		},
	}
}

// EdgeExternalIDKey is the payload key EdgeKeyLinks reads the external ids schema's edge links to from
func EdgeExternalIDKey(schema *entityops.Schema, edge string) string {
	return externalIDMatch.payloadKey(schema, edge)
}

// EdgeNameKey is the payload key EdgeKeyLinks reads the names schema's edge links to from
func EdgeNameKey(schema *entityops.Schema, edge string) string {
	return nameMatch.payloadKey(schema, edge)
}

// EdgeExternalIDEntry binds a CEL expression producing the external ids schema's edge links to
func EdgeExternalIDEntry(schema *entityops.Schema, edge, expr string) entityops.MappingEntry {
	return entityops.MappingEntry{Key: EdgeExternalIDKey(schema, edge), Expr: expr}
}

// EdgeKeyLinks returns a link rule for every edge of schema through each match key its target supports, external id
// and name, reading the values from a payload key named after the edge: <edge>_external_id(s) and <edge>_name(s), the
// singular key for a unique edge and the plural key for a to-many edge. When edges are given only those edges are linked
func EdgeKeyLinks(schema *entityops.Schema, edges ...string) []types.LinkRule {
	return lo.FlatMap(edgeMatches, func(match edgeMatch, _ int) []types.LinkRule {
		return match.links(schema, edges)
	})
}

// key is the payload key carrying the match values for edge, singular for a unique edge and plural otherwise
func (m edgeMatch) key(edge entityops.EdgeDescriptor) string {
	return edge.Name + lo.Ternary(edge.Unique, m.one, m.many)
}

// payloadKey is the payload key carrying the match values for schema's named edge
func (m edgeMatch) payloadKey(schema *entityops.Schema, name string) string {
	edge, _ := schema.EdgeByName(name)
	edge.Name = name

	return m.key(edge)
}

// links returns the rules linking schema's edges through the match key, limited to edges when any are given
func (m edgeMatch) links(schema *entityops.Schema, edges []string) []types.LinkRule {
	return lo.FilterMap(schema.Edges, func(edge entityops.EdgeDescriptor, _ int) (types.LinkRule, bool) {
		if len(edges) > 0 && !lo.Contains(edges, edge.Name) {
			return types.LinkRule{}, false
		}

		if edge.Target == nil || edge.Target.QueryByKey == nil || !edge.Target.MatchKeyField(m.field) {
			return types.LinkRule{}, false
		}

		key := m.key(edge)
		if inputKeyCollides(schema, key) {
			return types.LinkRule{}, false
		}

		rule := types.LinkRule{TargetSchema: edge.TargetType, Edge: edge.Name, TargetField: m.field, PayloadKeys: true}
		if edge.Unique {
			rule.SourceField = key
		} else {
			rule.SourceList = key
		}

		return rule, true
	})
}

// inputKeyCollides reports whether key is already an input field or edge create key on schema
func inputKeyCollides(schema *entityops.Schema, key string) bool {
	return lo.ContainsBy(schema.Fields, func(f entityops.FieldDescriptor) bool { return f.InputKey == key }) ||
		lo.ContainsBy(schema.Edges, func(e entityops.EdgeDescriptor) bool { return e.CreateField == key })
}

// internalOwnerIdentityHolderLink links the internal owner identity holder whose email equals the mapped internal_owner value
func internalOwnerIdentityHolderLink(edge, sourceKey string) types.LinkRule {
	return types.LinkRule{
		TargetSchema: entityops.SchemaIdentityHolder.Name,
		Edge:         edge,
		TargetField:  identityholder.FieldEmail,
		SourceField:  sourceKey,
	}
}
