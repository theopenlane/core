package googleworkspace

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	admin "google.golang.org/api/admin/directory/v1"

	"github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// directorySyncRequest builds an OperationRequest carrying the given installation metadata attributes
func directorySyncRequest(attributes string) types.OperationRequest {
	return types.OperationRequest{
		Integration: &ent.Integration{
			ID: "install-1",
			InstallationMetadata: openapi.IntegrationInstallationMetadata{
				Attributes: json.RawMessage(attributes),
			},
		},
		Client: &admin.Service{},
	}
}

func TestDirectorySyncMissingCustomerIDIsUnhealthy(t *testing.T) {
	t.Parallel()

	_, err := DirectorySync{}.IngestHandle()(context.Background(), directorySyncRequest(`{"domain":"example.com"}`))
	if !errors.Is(err, ErrCustomerIDMissing) {
		t.Fatalf("expected ErrCustomerIDMissing, got %v", err)
	}

	unhealthy, ok := types.UnhealthyFrom(err)
	if !ok {
		t.Fatal("expected a missing customer id to be a terminal unhealthy failure")
	}

	if unhealthy.Reason == "" {
		t.Fatal("expected a user-facing reason on the unhealthy error")
	}
}

func TestDirectorySyncInvalidMetadataIsNotUnhealthy(t *testing.T) {
	t.Parallel()

	_, err := DirectorySync{}.IngestHandle()(context.Background(), directorySyncRequest(`{invalid`))
	if !errors.Is(err, ErrInstallationMetadataInvalid) {
		t.Fatalf("expected ErrInstallationMetadataInvalid, got %v", err)
	}

	if _, ok := types.UnhealthyFrom(err); ok {
		t.Fatal("a metadata decode failure must not mark the integration unhealthy")
	}
}
