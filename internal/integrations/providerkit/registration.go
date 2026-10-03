package providerkit

import (
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated/control"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// matchAllFilterExpr is the filter expression admitting every envelope to its mapping
const matchAllFilterExpr = "true"

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
