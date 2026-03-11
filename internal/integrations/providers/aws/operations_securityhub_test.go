package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/internal/integrations/providers/awskit"
)

func TestSecurityHubFiltersFromMetadata_NoFilters(t *testing.T) {
	meta := awskit.Metadata{
		AccountScope: awskit.AccountScopeAll,
	}

	filters := securityHubFiltersFromMetadata(meta)
	assert.Nil(t, filters)
}

func TestSecurityHubFiltersFromMetadata_SpecificAccounts(t *testing.T) {
	meta := awskit.Metadata{
		AccountScope: awskit.AccountScopeSpecific,
		AccountIDs:   []string{"111111111111", "222222222222"},
	}

	filters := securityHubFiltersFromMetadata(meta)
	assert.NotNil(t, filters)
	assert.Len(t, filters.AwsAccountId, 2)
}

func TestSecurityHubFiltersFromMetadata_LinkedRegions(t *testing.T) {
	meta := awskit.Metadata{
		AccountScope:  awskit.AccountScopeAll,
		LinkedRegions: []string{"us-east-1", "eu-west-1"},
	}

	filters := securityHubFiltersFromMetadata(meta)
	assert.NotNil(t, filters)
	assert.Len(t, filters.Region, 2)
}

func TestToSecurityHubStringFilters_FiltersEmpty(t *testing.T) {
	result := toSecurityHubStringFilters([]string{"", "", ""})
	assert.Nil(t, result)
}

func TestToSecurityHubStringFilters_Valid(t *testing.T) {
	result := toSecurityHubStringFilters([]string{"us-east-1", "eu-west-1"})
	assert.Len(t, result, 2)
}
