//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// oauthTokenCredV1 is the earlier OAuth credential shape without a refresh token
type oauthTokenCredV1 struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"access_token"`
}

// oauthTokenCredV2 is an OAuth credential shape carrying a refresh token
type oauthTokenCredV2 struct {
	// AccessToken is the OAuth2 access token
	AccessToken string `json:"access_token"`
	// RefreshToken is the OAuth2 refresh token
	RefreshToken string `json:"refresh_token,omitempty"`
}

// strictRegionInput is a user input layout requiring an unconverted region
type strictRegionInput struct {
	// Region is the required region
	Region string `json:"region" jsonschema:"required"`
}

// strictRegionInputRef is the strict region layout, replacing no earlier layout
var strictRegionInputRef = integrationtypes.UserInputRefOf[strictRegionInput]()

// previousOAuthDefinition returns an earlier version of the shared test definition
func previousOAuthDefinition(t *testing.T, current integrationtypes.Definition, schema json.RawMessage) registry.Builder {
	t.Helper()

	connection, err := current.ConnectionRegistration(testint.OAuthCredential.ID())
	require.NoError(t, err)

	return func() (integrationtypes.Definition, error) {
		return integrationtypes.Definition{
			DefinitionSpec: integrationtypes.DefinitionSpec{
				ID:          testint.DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
			CredentialRegistrations: []integrationtypes.CredentialRegistration{
				{Ref: testint.OAuthCredential.ID(), StoredSchema: schema},
			},
			HealthCheck: current.HealthCheck,
			Connections: []integrationtypes.ConnectionRegistration{
				{
					CredentialRef:  testint.OAuthCredential.ID(),
					CredentialRefs: []integrationtypes.CredentialSlotID{testint.OAuthCredential.ID()},
					Auth: &integrationtypes.AuthRegistration{
						CredentialRef: testint.OAuthCredential.ID(),
						Start:         connection.Auth.Start,
						Complete:      connection.Auth.Complete,
					},
				},
			},
		}, nil
	}
}

func TestInstallationUpgradeEdges(t *testing.T) {
	store, err := keystore.NewStore(suite.Client.DB)
	require.NoError(t, err)

	def, ok := suite.IntegrationsRT.Registry().Definition(testint.DefinitionID.ID())
	require.True(t, ok)

	current := suite.IntegrationsRT.Registry().Version(def.ID)
	require.NotEmpty(t, current)

	t.Run("an auth-managed slot is conformed against the auth flow schema on upgrade", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		installation, previous := installUnder(t, subCtx, previousOAuthDefinition(t, def, jsonx.SchemaFrom[oauthTokenCredV1]()), testint.OAuthCredential.ID(), map[integrationtypes.CredentialSlotID]integrationtypes.CredentialSet{
			testint.OAuthCredential.ID(): {Data: json.RawMessage(`{"access_token":"legacy-oauth-token","legacy":"drop-me"}`)},
		})
		require.NotEmpty(t, previous)
		require.NotEqual(t, current, previous)
		require.Equal(t, previous, installation.DefinitionVersion)

		assessment, err := suite.IntegrationsRT.RunHealthAssessment(subCtx, installation)
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)

		rows, err := store.LoadAllCredentials(subCtx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"access_token":"legacy-oauth-token"}`, string(rows[testint.OAuthCredential.ID()].Data))

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, current, reloaded.DefinitionVersion)

		state, err := def.ProviderState(reloaded.ProviderState)
		require.NoError(t, err)
		require.Equal(t, testint.OAuthCredential.ID(), state.CredentialRef)

		registration, err := def.CredentialRegistration(testint.OAuthCredential.ID())
		require.NoError(t, err)
		require.NotEmpty(t, registration.StoredSchema)
		require.Empty(t, registration.Schema)
	})

	t.Run("a user input the current schema rejects with no conversion fails the upgrade and marks the installation errored", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.UserInput = zoneInputRef.Registration()
		}))
		installation := installOn(t, subCtx, previous, json.RawMessage(`{"zone":"eu"}`), nil, testint.TokenCredential.ID(), testint.TokenCredentialSet("token"))
		require.Equal(t, zoneInputRef.Name(), installation.UserInput.Layout)
		require.JSONEq(t, `{"zone":"eu"}`, string(installation.UserInput.Data))

		strict := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.UserInput = strictRegionInputRef.Registration()
		}))
		strictVersion := strict.Registry().Version(testint.DefinitionID.ID())
		require.NotEqual(t, strictVersion, installation.DefinitionVersion)

		_, err := strict.RunHealthAssessment(subCtx, installation)
		require.ErrorIs(t, err, intruntime.ErrInstallationUpgradeFailed)
		require.ErrorIs(t, err, intruntime.ErrUserInputInvalid)

		_, unhealthy := integrationtypes.UnhealthyFrom(err)
		require.True(t, unhealthy)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, installation.DefinitionVersion, reloaded.DefinitionVersion)
		require.Equal(t, enums.IntegrationStatusErrored, reloaded.Status)
		require.Equal(t, zoneInputRef.Name(), reloaded.UserInput.Layout)
		require.JSONEq(t, `{"zone":"eu"}`, string(reloaded.UserInput.Data))
	})

	t.Run("the auth-managed slot's schema participates in the version hash", func(t *testing.T) {
		v1 := runtimeFor(t, previousOAuthDefinition(t, def, jsonx.SchemaFrom[oauthTokenCredV1]()))
		v2 := runtimeFor(t, previousOAuthDefinition(t, def, jsonx.SchemaFrom[oauthTokenCredV2]()))

		require.NotEqual(t, v1.Registry().Version(testint.DefinitionID.ID()), v2.Registry().Version(testint.DefinitionID.ID()))
	})
}
