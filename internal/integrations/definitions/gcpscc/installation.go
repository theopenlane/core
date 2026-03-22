package gcpscc

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/internal/integrations/types"
)

type serviceAccountIdentity struct {
	ClientEmail string `json:"client_email"`
}

// resolveInstallationMetadata derives GCP SCC installation metadata from the persisted credential
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	meta, err := metadataFromCredential(req.Credential)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	if meta.OrganizationID == "" && meta.ProjectID == "" && meta.SourceID == "" && len(meta.SourceIDs) == 0 {
		return InstallationMetadata{}, false, nil
	}

	var serviceAccount serviceAccountIdentity
	if key := normalizeServiceAccountKey(meta.ServiceAccountKey); key != "" {
		_ = json.Unmarshal([]byte(key), &serviceAccount)
	}

	return InstallationMetadata{
		OrganizationID:      meta.OrganizationID,
		ProjectID:           meta.ProjectID,
		ProjectScope:        meta.ProjectScope,
		ProjectIDs:          meta.ProjectIDs,
		SourceID:            meta.SourceID,
		SourceIDs:           meta.SourceIDs,
		ServiceAccountEmail: serviceAccount.ClientEmail,
	}, true, nil
}
