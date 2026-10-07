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
func previousOAuthDefinition[T any](t *testing.T, current integrationtypes.Definition) registry.Builder {
	t.Helper()

	connection, ok := current.Connection(testint.OAuth.Connection().Credential.Name)
	require.True(t, ok)

	flow := integrationtypes.NewAuthFlow[T](connection.Auth.Start, connection.Auth.Complete)

	return func() (integrationtypes.Definition, error) {
		return integrationtypes.Definition{
			DefinitionSpec: integrationtypes.DefinitionSpec{
				ID:          testint.DefinitionID.ID(),
				DisplayName: "Test Integration",
				Active:      true,
			},
			Installation: integrationtypes.InstallationOf[testMetadata]().Registration(),
			Connections: []integrationtypes.Connector{
				slotOf(integrationtypes.NewConnection[T](connection.Credential.Name).Authenticates(flow)).connector,
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

		installation, previous := installUnder(t, subCtx, previousOAuthDefinition[oauthTokenCredV1](t, def), testint.OAuth.Connection().Credential.Name, map[string]integrationtypes.CredentialSet{
			testint.OAuth.Connection().Credential.Name: {Data: json.RawMessage(`{"access_token":"legacy-oauth-token","legacy":"drop-me"}`)},
		})
		require.Less(t, previous, current)
		require.Equal(t, previous, installation.DefinitionVersion)

		assessment, err := suite.IntegrationsRT.RunHealthAssessment(subCtx, installation)
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)

		rows, err := store.LoadAllCredentials(subCtx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"access_token":"legacy-oauth-token"}`, string(rows[testint.OAuth.Connection().Credential.Name].Data))

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, current, reloaded.DefinitionVersion)

		state, err := def.ProviderState(reloaded.ProviderState)
		require.NoError(t, err)
		require.Equal(t, testint.OAuth.Connection().Credential.Name, state.CredentialRef)

		connection, ok := def.Connection(testint.OAuth.Connection().Credential.Name)
		require.True(t, ok)
		require.NotEmpty(t, connection.Credential.Schema)
		require.Empty(t, connection.Form)
	})

	t.Run("a user input the current schema rejects with no conversion fails the upgrade and marks the installation errored", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.UserInput = zoneInputRef.Registration()
		}))
		installation := installOn(t, subCtx, previous, json.RawMessage(`{"zone":"eu"}`), nil, testint.Token.Connection().Credential.Name, testint.TokenCredentialSet("token"))
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

	t.Run("corrected user input supplied to reconcile repairs a stranded installation in place and completes the upgrade", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.UserInput = zoneInputRef.Registration()
		}))
		installation := installOn(t, subCtx, previous, json.RawMessage(`{"zone":"eu"}`), nil, testint.Token.Connection().Credential.Name, testint.TokenCredentialSet("token"))

		strict := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.UserInput = strictRegionInputRef.Registration()
		}))
		strictVersion := strict.Registry().Version(testint.DefinitionID.ID())

		_, err := strict.RunHealthAssessment(subCtx, installation)
		require.ErrorIs(t, err, intruntime.ErrInstallationUpgradeFailed)

		stranded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, enums.IntegrationStatusErrored, stranded.Status)
		require.NotEqual(t, strictVersion, stranded.DefinitionVersion)

		strictDef, ok := strict.Registry().Definition(testint.DefinitionID.ID())
		require.True(t, ok)

		_, _, err = strict.EnsureInstallation(subCtx, stranded.OwnerID, stranded.ID, strictDef, json.RawMessage(`{"region":"eu"}`), nil)
		require.NoError(t, err)

		repaired := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, strictRegionInputRef.Name(), repaired.UserInput.Layout)
		require.JSONEq(t, `{"region":"eu"}`, string(repaired.UserInput.Data))
		require.Equal(t, strictVersion, repaired.DefinitionVersion)
		require.Equal(t, enums.IntegrationStatusConnected, repaired.Status)

		rows, err := store.LoadAllCredentials(subCtx, repaired)
		require.NoError(t, err)
		require.JSONEq(t, `{"token":"token"}`, string(rows[testint.Token.Connection().Credential.Name].Data))
	})

	t.Run("operation config supplied to reconcile still upgrades every other stored document and completes the upgrade", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		previous := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.UserInput = zoneInputRef.Registration()
			def.Operations = []integrationtypes.OperationRegistration{syncOperation(retiredSyncOp, integrationtypes.ExecutionPolicy{Inline: true})}
		}))
		installation := installOn(t, subCtx, previous, json.RawMessage(`{"zone":"eu"}`), map[string]json.RawMessage{retiredSyncOp.Name(): json.RawMessage(`{"disable":true}`)}, testint.Token.Connection().Credential.Name, testint.TokenCredentialSet("token"))
		require.JSONEq(t, `{"disable":true}`, string(installation.OperationConfig.For(retiredSyncOp.Name())))

		renamed := runtimeFor(t, definitionOver(func(def *integrationtypes.Definition) {
			def.UserInput = regionInputRef.Registration()
			def.Operations = []integrationtypes.OperationRegistration{syncOperation(renamedSyncOp.Replacing(retiredSyncOp), integrationtypes.ExecutionPolicy{Inline: true})}
		}))
		renamedVersion := renamed.Registry().Version(testint.DefinitionID.ID())
		require.NotEqual(t, renamedVersion, installation.DefinitionVersion)

		renamedDef, ok := renamed.Registry().Definition(testint.DefinitionID.ID())
		require.True(t, ok)

		_, _, err := renamed.EnsureInstallation(subCtx, installation.OwnerID, installation.ID, renamedDef, nil, map[string]json.RawMessage{renamedSyncOp.Name(): json.RawMessage(`{"disable":false}`)})
		require.NoError(t, err)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, renamedVersion, reloaded.DefinitionVersion)
		require.Equal(t, regionInputRef.Name(), reloaded.UserInput.Layout)
		require.JSONEq(t, `{"region":"eu","token":"x"}`, string(reloaded.UserInput.Data))
		require.JSONEq(t, `{"disable":false}`, string(reloaded.OperationConfig.For(renamedSyncOp.Name())))
		require.Empty(t, reloaded.OperationConfig.For(retiredSyncOp.Name()))
		require.Equal(t, enums.IntegrationStatusConnected, reloaded.Status)
	})

	t.Run("the auth-managed slot's schema participates in the version hash", func(t *testing.T) {
		v1 := runtimeFor(t, previousOAuthDefinition[oauthTokenCredV1](t, def))
		v2 := runtimeFor(t, previousOAuthDefinition[oauthTokenCredV2](t, def))

		require.NotEqual(t, v1.Registry().Version(testint.DefinitionID.ID()), v2.Registry().Version(testint.DefinitionID.ID()))
	})
}
