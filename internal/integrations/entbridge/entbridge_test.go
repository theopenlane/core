package entbridge

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

func TestInstallationNil(t *testing.T) {
	t.Parallel()

	require.Nil(t, Installation(nil))
	require.Nil(t, Webhook(nil))
}

func TestInstallationCopiesFields(t *testing.T) {
	t.Parallel()

	record := &ent.Integration{
		ID:               "int-1",
		OwnerID:          "org-1",
		Name:             "Acme",
		DefinitionID:     "def-1",
		Status:           enums.IntegrationStatusConnected,
		CampaignEmail:    true,
		PrimaryDirectory: true,
		UserInput:        openapi.IntegrationUserInput{Data: json.RawMessage(`{"a":1}`)},
		InstallationMetadata: openapi.IntegrationInstallationMetadata{
			Attributes: json.RawMessage(`{"installationId":"42"}`),
		},
		Metadata: map[string]any{"k": "v"},
	}

	got := Installation(record)

	require.Equal(t, record.ID, got.ID)
	require.Equal(t, record.OwnerID, got.OwnerID)
	require.Equal(t, record.Name, got.Name)
	require.Equal(t, record.DefinitionID, got.DefinitionID)
	require.Equal(t, record.Status, got.Status)
	require.True(t, got.CampaignEmail)
	require.True(t, got.PrimaryDirectory)
	require.JSONEq(t, `{"a":1}`, string(got.UserInput.Data))
	require.JSONEq(t, `{"installationId":"42"}`, string(got.InstallationMetadata.Attributes))
	require.Equal(t, "v", got.Metadata["k"])
}

func TestWebhookCopiesFields(t *testing.T) {
	t.Parallel()

	url := "https://example.com/hook"
	got := Webhook(&ent.IntegrationWebhook{
		ID:            "wh-1",
		IntegrationID: "int-1",
		Name:          "events",
		EndpointURL:   &url,
		SecretToken:   "secret",
		AllowedEvents: []string{"push"},
	})

	require.Equal(t, "wh-1", got.ID)
	require.Equal(t, "int-1", got.IntegrationID)
	require.Equal(t, "events", got.Name)
	require.Equal(t, &url, got.EndpointURL)
	require.Equal(t, "secret", got.SecretToken)
	require.Equal(t, []string{"push"}, got.AllowedEvents)
}

func TestDBFromServices(t *testing.T) {
	t.Parallel()

	db := &ent.Client{}

	require.Same(t, db, DB(Services(db)))

	got, err := RequireDB(Services(db))
	require.NoError(t, err)
	require.Same(t, db, got)

	require.Nil(t, DB(nil))

	_, err = RequireDB(nil)
	require.ErrorIs(t, err, ErrNoDB)
}

func TestDBOnlyServicesRejectRuntimeCapabilities(t *testing.T) {
	t.Parallel()

	services := Services(&ent.Client{})

	_, err := services.Dispatch(context.Background(), types.DispatchRequest{})
	require.ErrorIs(t, err, errServiceUnavailable)

	_, err = services.ResolveOwnerIntegration(context.Background(), "def", "org")
	require.ErrorIs(t, err, errServiceUnavailable)

	_, found := services.Definitions().Definition("def")
	require.False(t, found)
}

func TestRecordPrefersRuntimeInstance(t *testing.T) {
	t.Parallel()

	record := &ent.Integration{ID: "int-1"}
	ctx := WithRecord(context.Background(), record)

	got, err := Record(ctx, nil, &types.Installation{ID: "int-1"})
	require.NoError(t, err)
	require.Same(t, record, got)

	got, err = Record(ctx, nil, nil)
	require.NoError(t, err)
	require.Nil(t, got)
}
