package gcpscc

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/types"
)

// serviceAccountIdentity represents the identity fields extracted from a GCP service account key
type serviceAccountIdentity struct {
	// ClientEmail is the email address of the GCP service account
	ClientEmail string `json:"client_email"`
}

// resolveInstallationMetadata derives GCP SCC installation metadata from the persisted credential
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	scope, err := resolveScope(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	if scope.OrganizationID == "" && scope.ProjectID == "" && len(scope.SourceIDs) == 0 {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		OrganizationID:      scope.OrganizationID,
		ProjectID:           scope.ProjectID,
		ProjectScope:        scope.ProjectScope,
		ProjectIDs:          scope.ProjectIDs,
		SourceIDs:           scope.SourceIDs,
		ServiceAccountEmail: serviceAccountEmail(req.Credentials),
	}, true, nil
}

// serviceAccountEmail returns the service account the installation acts as
func serviceAccountEmail(bindings types.CredentialBindings) string {
	federated, ok, err := workloadIdentityCredential.Resolve(bindings)
	if err == nil && ok {
		return federated.ServiceAccountEmail
	}

	cred, ok, err := sccCredential.Resolve(bindings)
	if err != nil || !ok {
		return ""
	}

	key := normalizeServiceAccountKey(cred.ServiceAccountKey)
	if key == "" {
		return ""
	}

	var identity serviceAccountIdentity

	_ = json.Unmarshal([]byte(key), &identity)

	return identity.ClientEmail
}
