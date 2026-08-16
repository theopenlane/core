package oidc

import (
	"github.com/samber/lo"
	zoidc "github.com/zitadel/oidc/v3/pkg/oidc"

	"github.com/theopenlane/iam/tokens"
)

// DiscoveryDocument builds the OIDC discovery document for the issuer
func DiscoveryDocument(manager *tokens.TokenManager) *zoidc.DiscoveryConfiguration {
	cfg := manager.Config()

	return &zoidc.DiscoveryConfiguration{
		Issuer:                           cfg.Issuer,
		JwksURI:                          cfg.JWKSEndpoint,
		ResponseTypesSupported:           []string{string(zoidc.ResponseTypeIDTokenOnly)},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: signingAlgorithms(manager),
	}
}

// signingAlgorithms returns the deduplicated algorithms of the active signing keys
func signingAlgorithms(manager *tokens.TokenManager) []string {
	return lo.Uniq(lo.FilterMap(manager.ListActiveKeys(), func(kid string, _ int) (string, bool) {
		metadata, err := manager.GetKeyMetadata(kid)
		if err != nil {
			return "", false
		}

		return metadata.Algorithm, true
	}))
}
