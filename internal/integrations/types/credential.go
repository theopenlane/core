package types //nolint:revive

import (
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/jsonx"
	"github.com/theopenlane/core/pkg/mapx"
)

// CloneCredentialSet returns a deep copy of a credential set.
func CloneCredentialSet(set models.CredentialSet) models.CredentialSet {
	cloned := set
	cloned.ProviderData = jsonx.CloneRawMessage(set.ProviderData)
	cloned.Claims = mapx.DeepCloneMapAny(set.Claims)

	if set.OAuthExpiry != nil {
		expiry := *set.OAuthExpiry
		cloned.OAuthExpiry = &expiry
	}

	return cloned
}

// InferAuthKind infers an auth kind from populated credential fields.
func InferAuthKind(set models.CredentialSet) AuthKind {
	switch {
	case len(set.Claims) > 0:
		return AuthKindOIDC
	case strings.TrimSpace(set.OAuthAccessToken) != "" || strings.TrimSpace(set.OAuthRefreshToken) != "":
		return AuthKindOAuth2
	case strings.TrimSpace(set.APIToken) != "":
		return AuthKindAPIKey
	case strings.TrimSpace(set.ClientID) != "" || strings.TrimSpace(set.ClientSecret) != "":
		return AuthKindOAuth2ClientCredentials
	case strings.TrimSpace(set.AccessKeyID) != "" ||
		strings.TrimSpace(set.SecretAccessKey) != "" ||
		strings.TrimSpace(set.SessionToken) != "":
		return AuthKindAWSFederation
	case strings.TrimSpace(set.ServiceAccountKey) != "" || strings.TrimSpace(set.SubjectToken) != "":
		return AuthKindWorkloadIdentity
	default:
		return AuthKindUnknown
	}
}

// IsCredentialSetEmpty reports whether all credential fields are empty.
func IsCredentialSetEmpty(set models.CredentialSet) bool {
	fields := []string{
		set.AccessKeyID,
		set.SecretAccessKey,
		set.SessionToken,
		set.ProjectID,
		set.AccountID,
		set.APIToken,
		set.ClientID,
		set.ClientSecret,
		set.ServiceAccountKey,
		set.SubjectToken,
		set.OAuthAccessToken,
		set.OAuthRefreshToken,
		set.OAuthTokenType,
	}

	if lo.ContainsBy(fields, func(field string) bool {
		return strings.TrimSpace(field) != ""
	}) {
		return false
	}

	if len(set.ProviderData) > 0 {
		return false
	}

	if set.OAuthExpiry != nil {
		return false
	}

	if len(set.Claims) > 0 {
		return false
	}

	return true
}

// MergeScopes merges scope values into a unique non-empty set.
func MergeScopes(dest []string, source ...string) []string {
	filtered := lo.Map(source, func(item string, _ int) string {
		return strings.TrimSpace(item)
	})

	nonEmpty := lo.Filter(filtered, func(item string, _ int) bool {
		return item != ""
	})

	return lo.Uniq(append(dest, nonEmpty...))
}
