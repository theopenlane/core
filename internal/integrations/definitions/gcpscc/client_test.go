package gcpscc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TestResolveCredential verifies credential resolution from bindings and default application
func TestResolveCredential(t *testing.T) {
	t.Run("decodes into credential schema", func(t *testing.T) {
		raw, err := jsonx.ToRawMessage(CredentialSchema{
			ServiceAccountKey: "\"{\\n  \\\"type\\\":\\\"service_account\\\"\\n}\"",
			CollectionScope: CollectionScope{
				ProjectID:      "project-123",
				OrganizationID: "org-123",
			},
		})
		require.NoError(t, err)

		bindings := types.CredentialBindings{
			{Ref: sccCredential.ID(), Credential: types.CredentialSet{Data: raw}},
		}

		meta, err := resolveCredential(bindings)
		require.NoError(t, err)

		assert.Equal(t, "project-123", meta.ProjectID)
		assert.Equal(t, "org-123", meta.OrganizationID)
		assert.Equal(t, "{\n  \"type\":\"service_account\"\n}", normalizeServiceAccountKey(meta.ServiceAccountKey))
	})

	t.Run("returns decode error for invalid provider data", func(t *testing.T) {
		bindings := types.CredentialBindings{
			{Ref: sccCredential.ID(), Credential: types.CredentialSet{Data: []byte(`{`)}},
		}

		_, err := resolveCredential(bindings)
		require.ErrorIs(t, err, ErrMetadataDecode)
	})

	t.Run("returns required error when binding is missing", func(t *testing.T) {
		_, err := resolveCredential(nil)
		require.ErrorIs(t, err, ErrCredentialMetadataRequired)
	})
}

// TestResolveScope verifies collection scope resolution from either credential slot
func TestResolveScope(t *testing.T) {
	t.Run("decodes collection scope from the workload identity slot", func(t *testing.T) {
		raw, err := jsonx.ToRawMessage(WorkloadIdentityCredentialSchema{
			ProjectNumber: "123456789",
			CollectionScope: CollectionScope{
				OrganizationID: "org-123",
			},
		})
		require.NoError(t, err)

		scope, err := resolveScope(types.CredentialBindings{
			{Ref: workloadIdentityCredential.ID(), Credential: types.CredentialSet{Data: raw}},
		})
		require.NoError(t, err)
		assert.Equal(t, "org-123", scope.OrganizationID)
	})

	t.Run("decodes collection scope from the service account slot", func(t *testing.T) {
		raw, err := jsonx.ToRawMessage(CredentialSchema{
			CollectionScope: CollectionScope{ProjectID: "project-123"},
		})
		require.NoError(t, err)

		scope, err := resolveScope(types.CredentialBindings{
			{Ref: sccCredential.ID(), Credential: types.CredentialSet{Data: raw}},
		})
		require.NoError(t, err)
		assert.Equal(t, "project-123", scope.ProjectID)
	})

	t.Run("returns required error when no slot is bound", func(t *testing.T) {
		_, err := resolveScope(nil)
		require.ErrorIs(t, err, ErrCredentialMetadataRequired)
	})
}
