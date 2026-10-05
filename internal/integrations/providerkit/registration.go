package providerkit

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// matchAllFilterExpr is the filter expression admitting every envelope to its mapping
const matchAllFilterExpr = "true"

// DirectorySync is the stored config of a directory sync operation that can skip groups and memberships;
// a directory without group sync declares its own DirectorySync embedding only types.OperationSettings
type DirectorySync struct {
	types.OperationSettings
	// DisableGroupSync skips collecting groups and memberships, syncing only users
	DisableGroupSync bool `json:"disableGroupSync,omitempty" jsonschema:"title=Disable Group Sync,description=Only sync users and skip group and membership sync"`
}

// UpgradeFromSection returns an upgrade decoding T from the object stored under key when present, else from the stored document;
// it reads main's client config section for an operation whose main key differs from its camelCase name
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

// FindingMapping returns the finding mapping linking findings to controls by ref code
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
			},
		},
	}
}
