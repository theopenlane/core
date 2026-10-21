package providerkit

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/asset"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/v2/internal/ent/generated/vulnerability"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/utils/jsonx"
)

// matchAllFilterExpr is the filter expression admitting every envelope to its mapping
const matchAllFilterExpr = "true"

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

// FindingMapping returns the finding mapping linking findings to controls by ref code, assets by name, and the internal owner identity holder by email
func FindingMapping(mapExpr string) types.MappingRegistration {
	return types.MappingRegistration{
		Schema: entityops.SchemaFinding.Name,
		Spec: types.MappingOverride{
			FilterExpr: matchAllFilterExpr,
			MapExpr:    mapExpr,
			Links: []types.LinkRule{
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
			},
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

// VulnerabilityMapping returns the vulnerability mapping linking the internal owner to an identity holder by email
func VulnerabilityMapping(mapExpr string) types.MappingRegistration {
	return types.MappingRegistration{
		Schema: entityops.SchemaVulnerability.Name,
		Spec: types.MappingOverride{
			FilterExpr: matchAllFilterExpr,
			MapExpr:    mapExpr,
			Links: []types.LinkRule{
				internalOwnerIdentityHolderLink(vulnerability.EdgeInternalOwnerIdentityHolder, entityops.VulnerabilityFields.InternalOwner.InputKey),
			},
		},
	}
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
