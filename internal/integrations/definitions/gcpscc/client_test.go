package gcpscc

import (
	"context"
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// TestServiceAccountCredentials verifies service account key validation
func TestServiceAccountCredentials(t *testing.T) {
	t.Run("rejects an empty key", func(t *testing.T) {
		_, err := serviceAccountCredentials(context.Background(), "  ")
		assert.ErrorIs(t, err, ErrServiceAccountKeyInvalid)
	})

	t.Run("rejects a malformed key", func(t *testing.T) {
		_, err := serviceAccountCredentials(context.Background(), "{")
		assert.ErrorIs(t, err, ErrServiceAccountKeyInvalid)
	})
}

// TestServiceAccountClient verifies the service account client rejects an unusable key before building
func TestServiceAccountClient(t *testing.T) {
	_, err := serviceAccountClient(context.Background(), types.ConnectionRequest[CredentialSchema]{
		Integration: &ent.Integration{},
		Credential:  CredentialSchema{ServiceAccountKey: ""},
	})
	assert.ErrorIs(t, err, ErrServiceAccountKeyInvalid)
}

// TestWorkloadIdentityClient verifies the workload identity client requires a project number before building
func TestWorkloadIdentityClient(t *testing.T) {
	_, err := workloadIdentityClient(context.Background(), types.ConnectionRequest[WorkloadIdentityCredentialSchema]{
		Integration: &ent.Integration{},
	})
	assert.ErrorIs(t, err, ErrProjectNumberRequired)
}

// TestKeyClientEmail verifies the client email is read from a service account key
func TestKeyClientEmail(t *testing.T) {
	t.Run("reads the client email from a raw key", func(t *testing.T) {
		assert.Check(t, is.Equal("collector@project-123.iam.gserviceaccount.com", keyClientEmail(`{"client_email":"collector@project-123.iam.gserviceaccount.com"}`)))
	})

	t.Run("reads the client email from a json-encoded key", func(t *testing.T) {
		assert.Check(t, is.Equal("collector@project-123.iam.gserviceaccount.com", keyClientEmail(`"{\"client_email\":\"collector@project-123.iam.gserviceaccount.com\"}"`)))
	})

	t.Run("returns empty for an empty or malformed key", func(t *testing.T) {
		assert.Check(t, is.Len(keyClientEmail(""), 0))
		assert.Check(t, is.Len(keyClientEmail("{"), 0))
	})
}
