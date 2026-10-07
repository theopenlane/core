package gcpscc

import (
	"context"
	"encoding/json"
	"errors"

	securitycenterpb "cloud.google.com/go/securitycenter/apiv2/securitycenterpb"
	"google.golang.org/api/iterator"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/logx"
)

// serviceAccountIdentity represents the identity fields extracted from a GCP service account key
type serviceAccountIdentity struct {
	// ClientEmail is the email address of the GCP service account
	ClientEmail string `json:"client_email"`
}

// verifyWorkloadIdentity probes SCC through the workload identity client and derives the installation metadata
func verifyWorkloadIdentity(ctx context.Context, req types.ConnectionRequest[WorkloadIdentityCredentialSchema], c Client) (InstallationMetadata, error) {
	if _, err := probeSources(ctx, c); err != nil {
		return InstallationMetadata{}, err
	}

	return InstallationMetadata{
		OrganizationID:      c.Scope.OrganizationID,
		ProjectID:           c.Scope.ProjectID,
		ProjectScope:        c.Scope.ProjectScope,
		ProjectIDs:          c.Scope.ProjectIDs,
		SourceIDs:           c.Scope.SourceIDs,
		ServiceAccountEmail: req.Credential.ServiceAccountEmail,
	}, nil
}

// verifyServiceAccount probes SCC through the service account client and derives the installation metadata
func verifyServiceAccount(ctx context.Context, req types.ConnectionRequest[CredentialSchema], c Client) (InstallationMetadata, error) {
	if _, err := probeSources(ctx, c); err != nil {
		return InstallationMetadata{}, err
	}

	return InstallationMetadata{
		OrganizationID:      c.Scope.OrganizationID,
		ProjectID:           c.Scope.ProjectID,
		ProjectScope:        c.Scope.ProjectScope,
		ProjectIDs:          c.Scope.ProjectIDs,
		SourceIDs:           c.Scope.SourceIDs,
		ServiceAccountEmail: keyClientEmail(req.Credential.ServiceAccountKey),
	}, nil
}

// probeSources lists one source under each parent of the client's scope and returns the parents
func probeSources(ctx context.Context, c Client) ([]string, error) {
	parents, err := resolveParents(c.Scope)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("gcpscc: error attempting to resolve parents")
		return nil, err
	}

	for _, parent := range parents {
		it := c.ListSources(ctx, &securitycenterpb.ListSourcesRequest{
			Parent:   parent,
			PageSize: 1,
		})

		_, err = it.Next()

		if errors.Is(err, iterator.Done) {
			err = nil
		}

		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("gcpscc: failed to list sources")
			return nil, ErrListSourcesFailed
		}
	}

	return parents, nil
}

// keyClientEmail returns the client_email of a service account key
func keyClientEmail(rawKey string) string {
	key := normalizeServiceAccountKey(rawKey)
	if key == "" {
		return ""
	}

	var identity serviceAccountIdentity

	_ = json.Unmarshal([]byte(key), &identity)

	return identity.ClientEmail
}
