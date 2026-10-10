package gcpscc

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/v2/internal/integrations/auth"
)

const (
	// googleSTSEndpoint is the Google Security Token Service RFC 8693 exchange endpoint
	googleSTSEndpoint = "https://sts.googleapis.com/v1/token"
	// workloadIdentityName is the fixed pool and provider ID customers create for Openlane federation
	workloadIdentityName = "openlane"
	// workloadIdentityAudienceFormat is the provider resource name Google STS expects as the exchange audience
	workloadIdentityAudienceFormat = "//iam.googleapis.com/projects/%s/locations/global/workloadIdentityPools/%s/providers/%s"
)

// workloadIdentityAudience builds the STS exchange audience for the pool-hosting project number
func workloadIdentityAudience(projectNumber string) string {
	return fmt.Sprintf(workloadIdentityAudienceFormat, projectNumber, workloadIdentityName, workloadIdentityName)
}

// federationSource builds the token source, impersonating a service account when configured
func federationSource(ctx context.Context, manager *tokens.TokenManager, organizationID string, cred WorkloadIdentityCredentialSchema) (oauth2.TokenSource, error) {
	if cred.ProjectNumber == "" {
		return nil, ErrProjectNumberRequired
	}

	federated, err := auth.FederatedTokenSource(ctx, manager, organizationID, auth.FederationSpec{
		Audience: workloadIdentityAudience(cred.ProjectNumber),
		Scopes:   []string{defaultScope},
		Endpoint: googleSTSEndpoint,
	})
	if err != nil {
		return nil, err
	}

	if cred.ServiceAccountEmail == "" {
		return federated, nil
	}

	return impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: cred.ServiceAccountEmail,
		Scopes:          []string{defaultScope},
	}, option.WithTokenSource(federated))
}
