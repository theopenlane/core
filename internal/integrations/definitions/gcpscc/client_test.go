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
	t.Run("decodes into credential schema and applies defaults", func(t *testing.T) {
		raw, err := jsonx.ToRawMessage(CredentialSchema{
			ProjectID:         "project-123",
			OrganizationID:    "org-123",
			ServiceAccountKey: "\"{\\n  \\\"type\\\":\\\"service_account\\\"\\n}\"",
		})
		require.NoError(t, err)

		bindings := types.CredentialBindings{
			{Ref: sccCredential.ID(), Credential: types.CredentialSet{Data: raw}},
		}

		meta, err := resolveCredential(bindings)
		require.NoError(t, err)

		assert.Equal(t, "project-123", meta.ProjectID)
		assert.Equal(t, "org-123", meta.OrganizationID)
		assert.Equal(t, projectScopeAll, meta.ProjectScope)
		assert.Equal(t, "{\n  \"type\":\"service_account\"\n}", meta.ServiceAccountKey)
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
