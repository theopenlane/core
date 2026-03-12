package types

import (
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
)

// CredentialSet is an alias for models.CredentialSet
type CredentialSet = models.CredentialSet

// IsCredentialSetEmpty reports whether all fields in the CredentialSet are empty or nil
func IsCredentialSetEmpty(set CredentialSet) bool {
	return set.OAuthAccessToken == "" &&
		set.OAuthRefreshToken == "" &&
		set.OAuthTokenType == "" &&
		set.OAuthExpiry == nil &&
		len(set.ProviderData) == 0 &&
		len(set.Claims) == 0
}

// CloneCredentialSet returns a deep copy of a CredentialSet
func CloneCredentialSet(set CredentialSet) CredentialSet {
	cloned := set
	cloned.ProviderData = jsonx.CloneRawMessage(set.ProviderData)
	cloned.Claims = mapx.DeepCloneMapAny(set.Claims)

	if set.OAuthExpiry != nil {
		expiry := *set.OAuthExpiry
		cloned.OAuthExpiry = &expiry
	}

	return cloned
}
