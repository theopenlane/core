package oci

import (
	"context"

	"github.com/theopenlane/core/internal/integrations/types"
)

// resolveInstallationMetadata derives OCI tenancy metadata from the persisted credential
func resolveInstallationMetadata(_ context.Context, req types.InstallationRequest) (InstallationMetadata, bool, error) {
	meta, err := resolveCredential(req.Credentials)
	if err != nil {
		return InstallationMetadata{}, false, err
	}

	if meta.TenancyOCID == "" {
		return InstallationMetadata{}, false, nil
	}

	return InstallationMetadata{
		TenancyOCID:     meta.TenancyOCID,
		CompartmentOCID: meta.CompartmentOCID,
		Region:          meta.Region,
	}, true, nil
}
