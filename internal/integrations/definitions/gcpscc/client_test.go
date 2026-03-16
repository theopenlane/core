package gcpscc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TestMetadataFromCredential verifies credential metadata decoding and default application
func TestMetadataFromCredential(t *testing.T) {
	t.Run("decodes into credential schema and applies defaults", func(t *testing.T) {
		raw, err := jsonx.ToRawMessage(CredentialSchema{
			ProjectID:         "project-123",
			OrganizationID:    "org-123",
			SourceID:          "source-123",
			FindingFilter:     "state=\"ACTIVE\"",
			ServiceAccountKey: "\"{\\n  \\\"type\\\":\\\"service_account\\\"\\n}\"",
		})
		require.NoError(t, err)

		meta, err := metadataFromCredential(types.CredentialSet{ProviderData: raw})
		require.NoError(t, err)

		assert.Equal(t, "project-123", meta.ProjectID)
		assert.Equal(t, "org-123", meta.OrganizationID)
		assert.Equal(t, "source-123", meta.SourceID)
		assert.Equal(t, projectScopeAll, meta.ProjectScope)
		assert.Equal(t, "{\n  \"type\":\"service_account\"\n}", meta.ServiceAccountKey)
	})

	t.Run("returns decode error for invalid provider data", func(t *testing.T) {
		_, err := metadataFromCredential(types.CredentialSet{ProviderData: []byte(`{`)})
		require.ErrorIs(t, err, ErrMetadataDecode)
	})
}
