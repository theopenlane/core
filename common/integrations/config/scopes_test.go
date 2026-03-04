package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/integrations/config"
)

func TestMergeRequestedScopes_NilOAuth(t *testing.T) {
	spec := config.ProviderSpec{}
	result := config.MergeRequestedScopes(spec, nil)
	assert.Nil(t, result)
}

func TestMergeRequestedScopes_EmptyBoth(t *testing.T) {
	spec := config.ProviderSpec{OAuth: &config.OAuthSpec{Scopes: []string{}}}
	result := config.MergeRequestedScopes(spec, []string{})
	assert.Nil(t, result)
}

func TestMergeRequestedScopes_SpecScopesOnly(t *testing.T) {
	spec := config.ProviderSpec{
		OAuth: &config.OAuthSpec{Scopes: []string{"read:org", "repo"}},
	}
	result := config.MergeRequestedScopes(spec, nil)
	assert.ElementsMatch(t, []string{"read:org", "repo"}, result)
}

func TestMergeRequestedScopes_RequestedOnly(t *testing.T) {
	spec := config.ProviderSpec{}
	result := config.MergeRequestedScopes(spec, []string{"email", "profile"})
	assert.ElementsMatch(t, []string{"email", "profile"}, result)
}

func TestMergeRequestedScopes_MergesWithDedup(t *testing.T) {
	spec := config.ProviderSpec{
		OAuth: &config.OAuthSpec{Scopes: []string{"read:org", "repo"}},
	}
	result := config.MergeRequestedScopes(spec, []string{"email", "repo"})
	assert.ElementsMatch(t, []string{"read:org", "repo", "email"}, result)
}

func TestMergeRequestedScopes_CaseSensitiveDedup(t *testing.T) {
	spec := config.ProviderSpec{
		OAuth: &config.OAuthSpec{Scopes: []string{"Read:Org"}},
	}
	result := config.MergeRequestedScopes(spec, []string{"read:org"})
	assert.Len(t, result, 2)
	assert.ElementsMatch(t, []string{"Read:Org", "read:org"}, result)
}

func TestMergeRequestedScopes_BlankRequestedScopesSkipped(t *testing.T) {
	spec := config.ProviderSpec{}
	result := config.MergeRequestedScopes(spec, []string{"  ", "", "email"})
	assert.ElementsMatch(t, []string{"email"}, result)
}
