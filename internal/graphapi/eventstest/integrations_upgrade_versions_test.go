//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationrun"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// TestInstallationUpgradeAcrossVersions drives the testutils/integrations version fixtures through real multi-version upgrades, including a direct v1-to-v3 jump
func TestInstallationUpgradeAcrossVersions(t *testing.T) {
	store, err := keystore.NewStore(suite.Client.DB)
	require.NoError(t, err)

	t.Run("the version hash differs between every fixture version", func(t *testing.T) {
		v1 := runtimeFor(t, testint.BuilderV1())
		v2 := runtimeFor(t, testint.BuilderV2())
		v3 := runtimeFor(t, testint.BuilderV3())

		hashV1 := v1.Registry().Version(testint.DefinitionID.ID())
		hashV2 := v2.Registry().Version(testint.DefinitionID.ID())
		hashV3 := v3.Registry().Version(testint.DefinitionID.ID())

		require.NotEmpty(t, hashV1)
		require.NotEmpty(t, hashV2)
		require.NotEmpty(t, hashV3)
		require.NotEqual(t, hashV1, hashV2)
		require.NotEqual(t, hashV2, hashV3)
		require.NotEqual(t, hashV1, hashV3)
	})

	t.Run("an installation walks v1 to v2 to v3 with every stored structure transformed", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		v1 := runtimeFor(t, testint.BuilderV1())
		installation := installOn(t, subCtx, v1, json.RawMessage(`{"filterExpr":"initial"}`), testint.TokenV1.ID(), testint.TokenV1Set("initial-token"))
		require.Equal(t, v1.Registry().Version(testint.DefinitionID.ID()), installation.DefinitionVersion)

		v2 := runtimeFor(t, testint.BuilderV2())
		v2Version := v2.Registry().Version(testint.DefinitionID.ID())

		assessment, err := v2.RunHealthAssessment(subCtx, installation)
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, v2Version, reloaded.DefinitionVersion)
		require.JSONEq(t, `{"filterExpr":"initial"}`, string(reloaded.Config.ClientConfig))

		rows, err := store.LoadAllCredentials(subCtx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"token":"initial-token"}`, string(rows[testint.TokenV2.ID()].Data))
		require.Empty(t, slotRowIDs(t, subCtx, installation.ID, testint.TokenV1.ID()))

		retired, err := v2.Registry().Operation(testint.DefinitionID.ID(), testint.SyncOp.Name())
		require.NoError(t, err)

		finished, err := operations.CreatePendingRun(subCtx, suite.Client.DB, installation, retired, enums.IntegrationRunTypeManual, nil)
		require.NoError(t, err)
		require.NoError(t, operations.CompleteRun(subCtx, suite.Client.DB, finished.ID, finished.StartedAt, operations.RunResult{}))

		finishedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, testint.SyncOp.Name())
		require.NoError(t, err)
		require.NotNil(t, finishedAt)

		require.NoError(t, suite.Client.DB.Integration.UpdateOneID(installation.ID).
			SetHealth(models.IntegrationHealth{UnhealthyOperations: map[string]string{testint.SyncOp.Name(): "boom"}}).
			Exec(subCtx))

		webhookRows := endpointRows(t, subCtx, installation.ID)
		require.Len(t, webhookRows, 1)
		require.Equal(t, testint.WebhookV1V2.Name(), webhookRows[0].Name)
		endpointID := lo.FromPtr(webhookRows[0].EndpointID)
		secret := webhookRows[0].SecretToken
		require.NotEmpty(t, endpointID)

		v3 := runtimeFor(t, testint.BuilderV3())
		v3Version := v3.Registry().Version(testint.DefinitionID.ID())

		assessment, err = v3.RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)
		require.Equal(t, []intruntime.OperationHealthResult{{Name: testint.SyncOpV3.Name(), Reason: "boom"}}, assessment.Operations)

		reloaded = reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, v3Version, reloaded.DefinitionVersion)
		require.JSONEq(t, `{"filter":"initial"}`, string(reloaded.Config.ClientConfig))
		require.Equal(t, map[string]string{testint.SyncOpV3.Name(): "boom"}, reloaded.Health.UnhealthyOperations)

		rows, err = store.LoadAllCredentials(subCtx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"token":"initial-token","region":"`+installation.ID+`"}`, string(rows[testint.TokenV3.ID()].Data))
		require.Empty(t, slotRowIDs(t, subCtx, installation.ID, testint.TokenV2.ID()))

		names, err := suite.Client.DB.IntegrationRun.Query().
			Where(integrationrun.IntegrationIDEQ(installation.ID), integrationrun.OperationNameIn(testint.SyncOp.Name(), testint.SyncOpV3.Name())).
			Select(integrationrun.FieldOperationName).
			Strings(subCtx)
		require.NoError(t, err)
		require.Equal(t, []string{testint.SyncOpV3.Name()}, names)

		movedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, testint.SyncOpV3.Name())
		require.NoError(t, err)
		require.NotNil(t, movedAt)
		require.True(t, movedAt.Equal(*finishedAt))

		webhookRows = endpointRows(t, subCtx, installation.ID)
		require.Len(t, webhookRows, 1)
		require.Equal(t, testint.WebhookV3.Name(), webhookRows[0].Name)
		require.Equal(t, endpointID, lo.FromPtr(webhookRows[0].EndpointID))
		require.Equal(t, secret, webhookRows[0].SecretToken)
	})

	t.Run("an installation stored under v1 skips straight to v3", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		v1 := runtimeFor(t, testint.BuilderV1())
		installation := installOn(t, subCtx, v1, json.RawMessage(`{"filterExpr":"initial"}`), testint.TokenV1.ID(), testint.TokenV1Set("initial-token"))

		retired, err := v1.Registry().Operation(testint.DefinitionID.ID(), testint.SyncOp.Name())
		require.NoError(t, err)

		finished, err := operations.CreatePendingRun(subCtx, suite.Client.DB, installation, retired, enums.IntegrationRunTypeManual, nil)
		require.NoError(t, err)
		require.NoError(t, operations.CompleteRun(subCtx, suite.Client.DB, finished.ID, finished.StartedAt, operations.RunResult{}))

		finishedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, testint.SyncOp.Name())
		require.NoError(t, err)
		require.NotNil(t, finishedAt)

		require.NoError(t, suite.Client.DB.Integration.UpdateOneID(installation.ID).
			SetHealth(models.IntegrationHealth{UnhealthyOperations: map[string]string{testint.SyncOp.Name(): "boom"}}).
			Exec(subCtx))

		webhookRows := endpointRows(t, subCtx, installation.ID)
		require.Len(t, webhookRows, 1)
		endpointID := lo.FromPtr(webhookRows[0].EndpointID)
		secret := webhookRows[0].SecretToken

		v3 := runtimeFor(t, testint.BuilderV3())
		v3Version := v3.Registry().Version(testint.DefinitionID.ID())

		assessment, err := v3.RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		require.NoError(t, err)
		require.True(t, assessment.Connection.Healthy)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, v3Version, reloaded.DefinitionVersion)
		require.JSONEq(t, `{"filter":"initial"}`, string(reloaded.Config.ClientConfig))
		require.Equal(t, map[string]string{testint.SyncOpV3.Name(): "boom"}, reloaded.Health.UnhealthyOperations)

		rows, err := store.LoadAllCredentials(subCtx, installation)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.JSONEq(t, `{"token":"initial-token","region":"`+installation.ID+`"}`, string(rows[testint.TokenV3.ID()].Data))
		require.Empty(t, slotRowIDs(t, subCtx, installation.ID, testint.TokenV1.ID()))

		movedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, testint.SyncOpV3.Name())
		require.NoError(t, err)
		require.NotNil(t, movedAt)
		require.True(t, movedAt.Equal(*finishedAt))

		webhookRows = endpointRows(t, subCtx, installation.ID)
		require.Len(t, webhookRows, 1)
		require.Equal(t, testint.WebhookV3.Name(), webhookRows[0].Name)
		require.Equal(t, endpointID, lo.FromPtr(webhookRows[0].EndpointID))
		require.Equal(t, secret, webhookRows[0].SecretToken)
	})

	t.Run("an installation with no recorded version is upgraded and stamped on first use", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		v3 := runtimeFor(t, testint.BuilderV3())
		def, ok := v3.Registry().Definition(testint.DefinitionID.ID())
		require.True(t, ok)

		state, err := def.WithProviderState(integrationtypes.IntegrationProviderState{}, integrationtypes.DefinitionProviderState{CredentialRef: testint.TokenV3.ID()})
		require.NoError(t, err)

		installation, err := suite.Client.DB.Integration.Create().
			SetName("No Version").
			SetKind(testint.DefinitionID.ID()).
			SetDefinitionID(testint.DefinitionID.ID()).
			SetDefinitionVersion("").
			SetProviderState(state).
			SetStatus(enums.IntegrationStatusConnected).
			Save(subCtx)
		require.NoError(t, err)
		require.Empty(t, installation.DefinitionVersion)

		require.NoError(t, store.SaveCredential(subCtx, installation, testint.TokenV3.ID(), integrationtypes.CredentialSet{Data: json.RawMessage(`{"token":"incomplete-token"}`)}))

		v3Version := v3.Registry().Version(testint.DefinitionID.ID())

		_, err = v3.RunHealthAssessment(subCtx, installation)
		require.NoError(t, err)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		require.Equal(t, v3Version, reloaded.DefinitionVersion)

		rows, err := store.LoadAllCredentials(subCtx, installation)
		require.NoError(t, err)
		require.JSONEq(t, `{"token":"incomplete-token","region":"`+installation.ID+`"}`, string(rows[testint.TokenV3.ID()].Data))
	})

	t.Run("installation metadata is refreshed by the upgrade", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		v1 := runtimeFor(t, testint.BuilderV1())
		installation := installOn(t, subCtx, v1, json.RawMessage(`{"filterExpr":"x"}`), testint.TokenV1.ID(), testint.TokenV1Set("meta-token"))

		before := reloadIntegration(t, subCtx, installation.ID).InstallationMetadata
		require.Equal(t, installation.ID, before.Display.ExternalID)
		require.Empty(t, before.Attributes)

		v3 := runtimeFor(t, testint.BuilderV3())

		_, err := v3.RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		require.NoError(t, err)

		after := reloadIntegration(t, subCtx, installation.ID).InstallationMetadata
		require.NotEqual(t, before, after)
		require.Equal(t, installation.ID, after.Display.ExternalID)

		var attrs struct {
			Region string `json:"region"`
		}
		require.NoError(t, json.Unmarshal(after.Attributes, &attrs))
		require.Equal(t, installation.ID, attrs.Region)
	})

	t.Run("an operation executed with no explicit config resolves it from the installation input", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetContext(subOrg.UserCtx, suite.Client.DB)

		v3 := runtimeFor(t, testint.BuilderV3())
		installation := installOn(t, subCtx, v3, json.RawMessage(`{"filter":"only-mine"}`), testint.TokenV3.ID(), testint.TokenV3Set("token", "region-1"))

		syncOp, err := v3.Registry().Operation(testint.DefinitionID.ID(), testint.SyncOpV3.Name())
		require.NoError(t, err)

		_, err = v3.ExecuteOperation(subCtx, installation, syncOp, nil, nil)
		require.NoError(t, err)

		select {
		case raw := <-testint.SyncConfigV3:
			var cfg struct {
				Filter string `json:"filter"`
			}
			require.NoError(t, json.Unmarshal(raw, &cfg))
			require.NotEmpty(t, cfg.Filter)
			require.Equal(t, "only-mine", cfg.Filter)
		default:
			t.Fatal("expected the v3 reconcile handler to record its resolved config")
		}
	})
}
