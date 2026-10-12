//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/v2/internal/ent/generated/integrationrun"
	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	intruntime "github.com/theopenlane/core/v2/internal/integrations/runtime"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/keystore"
	testint "github.com/theopenlane/core/v2/internal/testutils/integrations"
)

// syncOpConfig returns the operation config storing the version 1 and 2 sync pattern
func syncOpConfig(pattern string) map[string]json.RawMessage {
	return map[string]json.RawMessage{testint.SyncOp.Name(): json.RawMessage(`{"pattern":"` + pattern + `"}`)}
}

// versionedDefinitionID keeps the versioned fixtures out of the suite registry so the harness status listener does not reseed the suite definition's loops onto their installations
var versionedDefinitionID = integrationtypes.NewDefinitionRef("def_01K0TESTVERSIONS0000000001")

// versionedRuntime returns an in-memory runtime running one fixture version registered under versionedDefinitionID
func versionedRuntime(t *testing.T, builder registry.Builder) *intruntime.Runtime {
	t.Helper()

	return runtimeFor(t, func() (integrationtypes.Definition, error) {
		def, err := builder()
		if err != nil {
			return integrationtypes.Definition{}, err
		}

		def.ID = versionedDefinitionID.ID()
		def.Operations = lo.Map(def.Operations, func(op integrationtypes.OperationRegistration, _ int) integrationtypes.OperationRegistration {
			op.Topic = versionedDefinitionID.OperationTopic(op.Name)

			return op
		})

		return def, nil
	})
}

// TestInstallationUpgradeAcrossVersions verifies multi-version upgrades including a v1-to-v3 jump
func TestInstallationUpgradeAcrossVersions(t *testing.T) {
	store, err := keystore.NewStore(suite.Client.DB)
	assert.NilError(t, err)

	t.Run("the version hash differs between every fixture version", func(t *testing.T) {
		v1 := versionedRuntime(t, testint.BuilderV1())
		v2 := versionedRuntime(t, testint.BuilderV2())
		v3 := versionedRuntime(t, testint.BuilderV3())

		hashV1 := v1.Registry().Version(versionedDefinitionID.ID())
		hashV2 := v2.Registry().Version(versionedDefinitionID.ID())
		hashV3 := v3.Registry().Version(versionedDefinitionID.ID())

		assert.Assert(t, hashV1 != "")
		assert.Assert(t, hashV2 != "")
		assert.Assert(t, hashV3 != "")
		assert.Assert(t, hashV1 != hashV2)
		assert.Assert(t, hashV2 != hashV3)
		assert.Assert(t, hashV1 != hashV3)
	})

	t.Run("an installation walks v1 to v2 to v3 with every stored structure transformed", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		v1 := versionedRuntime(t, testint.BuilderV1())
		installation := installOn(t, subCtx, v1, json.RawMessage(`{"filterExpr":"initial"}`), syncOpConfig("initial"), testint.TokenV1.Connection().Credential.Name, testint.TokenV1Set("initial-token"))
		assert.Equal(t, v1.Registry().Version(versionedDefinitionID.ID()), installation.DefinitionVersion)
		assert.Equal(t, testint.UserInputV1.Name(), installation.UserInput.Layout)
		assertJSONEqual(t, `{"pattern":"initial"}`, string(installation.OperationConfig.For(testint.SyncOp.Name())))

		v2 := versionedRuntime(t, testint.BuilderV2())
		v2Version := v2.Registry().Version(versionedDefinitionID.ID())

		assessment, err := v2.RunHealthAssessment(subCtx, installation)
		assert.NilError(t, err)
		assert.Assert(t, assessment.Connection.Healthy)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		assert.Equal(t, v2Version, reloaded.DefinitionVersion)
		assert.Equal(t, testint.UserInputV2.Name(), reloaded.UserInput.Layout)
		assertJSONEqual(t, `{"filterExpr":"initial"}`, string(reloaded.UserInput.Data))
		assertJSONEqual(t, `{"pattern":"initial"}`, string(reloaded.OperationConfig.For(testint.SyncOp.Name())))

		rows, err := store.LoadAllCredentials(subCtx, installation)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(rows, 1))
		assertJSONEqual(t, `{"token":"initial-token"}`, string(rows[testint.TokenV2.Connection().Credential.Name].Data))
		assert.Assert(t, is.Len(slotRowIDs(t, subCtx, installation.ID, testint.TokenV1.Connection().Credential.Name), 0))

		retired, err := v2.Registry().Operation(versionedDefinitionID.ID(), testint.SyncOp.Name())
		assert.NilError(t, err)

		finished, err := operations.CreatePendingRun(subCtx, suite.Client.DB, installation, retired, enums.IntegrationRunTypeManual, nil)
		assert.NilError(t, err)
		assert.NilError(t, operations.CompleteRun(subCtx, suite.Client.DB, finished.ID, finished.StartedAt, operations.RunResult{}))

		finishedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, testint.SyncOp.Name())
		assert.NilError(t, err)
		assert.Assert(t, finishedAt != nil)

		assert.NilError(t, suite.Client.DB.Integration.UpdateOneID(installation.ID).
			SetHealth(models.IntegrationHealth{UnhealthyOperations: map[string]string{testint.SyncOp.Name(): "boom"}}).
			Exec(subCtx))

		webhookRows := endpointRows(t, subCtx, installation.ID)
		assert.Assert(t, is.Len(webhookRows, 1))
		assert.Equal(t, testint.WebhookV1V2.Name(), webhookRows[0].Name)
		endpointID := lo.FromPtr(webhookRows[0].EndpointID)
		secret := webhookRows[0].SecretToken
		assert.Assert(t, endpointID != "")

		v3 := versionedRuntime(t, testint.BuilderV3())
		v3Version := v3.Registry().Version(versionedDefinitionID.ID())

		assessment, err = v3.RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		assert.NilError(t, err)
		assert.Assert(t, assessment.Connection.Healthy)
		assert.DeepEqual(t, []intruntime.OperationHealthResult{{Name: testint.SyncOpV3.Name(), Reason: "boom"}}, assessment.Operations)

		reloaded = reloadIntegration(t, subCtx, installation.ID)
		assert.Equal(t, v3Version, reloaded.DefinitionVersion)
		assert.Equal(t, testint.UserInputV3.Name(), reloaded.UserInput.Layout)
		assertJSONEqual(t, `{"filter":"initial"}`, string(reloaded.UserInput.Data))
		assertJSONEqual(t, `{"filter":"initial"}`, string(reloaded.OperationConfig.For(testint.SyncOpV3.Name())))
		assert.Assert(t, is.Nil(reloaded.OperationConfig.For(testint.SyncOp.Name())))
		assert.DeepEqual(t, map[string]string{testint.SyncOpV3.Name(): "boom"}, reloaded.Health.UnhealthyOperations)

		rows, err = store.LoadAllCredentials(subCtx, installation)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(rows, 1))
		assertJSONEqual(t, `{"token":"initial-token","region":"`+installation.ID+`"}`, string(rows[testint.TokenV3.Connection().Credential.Name].Data))
		assert.Assert(t, is.Len(slotRowIDs(t, subCtx, installation.ID, testint.TokenV2.Connection().Credential.Name), 0))

		names, err := suite.Client.DB.IntegrationRun.Query().
			Where(integrationrun.IntegrationIDEQ(installation.ID), integrationrun.OperationNameIn(testint.SyncOp.Name(), testint.SyncOpV3.Name())).
			Select(integrationrun.FieldOperationName).
			Strings(subCtx)
		assert.NilError(t, err)
		assert.DeepEqual(t, []string{testint.SyncOpV3.Name()}, names)

		movedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, testint.SyncOpV3.Name())
		assert.NilError(t, err)
		assert.Assert(t, movedAt != nil)
		assert.Assert(t, movedAt.Equal(*finishedAt))

		webhookRows = endpointRows(t, subCtx, installation.ID)
		assert.Assert(t, is.Len(webhookRows, 1))
		assert.Equal(t, testint.WebhookV3.Name(), webhookRows[0].Name)
		assert.Equal(t, endpointID, lo.FromPtr(webhookRows[0].EndpointID))
		assert.Equal(t, secret, webhookRows[0].SecretToken)
	})

	t.Run("an installation stored under v1 skips straight to v3", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		v1 := versionedRuntime(t, testint.BuilderV1())
		installation := installOn(t, subCtx, v1, json.RawMessage(`{"filterExpr":"initial"}`), syncOpConfig("initial"), testint.TokenV1.Connection().Credential.Name, testint.TokenV1Set("initial-token"))

		retired, err := v1.Registry().Operation(versionedDefinitionID.ID(), testint.SyncOp.Name())
		assert.NilError(t, err)

		finished, err := operations.CreatePendingRun(subCtx, suite.Client.DB, installation, retired, enums.IntegrationRunTypeManual, nil)
		assert.NilError(t, err)
		assert.NilError(t, operations.CompleteRun(subCtx, suite.Client.DB, finished.ID, finished.StartedAt, operations.RunResult{}))

		finishedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, testint.SyncOp.Name())
		assert.NilError(t, err)
		assert.Assert(t, finishedAt != nil)

		assert.NilError(t, suite.Client.DB.Integration.UpdateOneID(installation.ID).
			SetHealth(models.IntegrationHealth{UnhealthyOperations: map[string]string{testint.SyncOp.Name(): "boom"}}).
			Exec(subCtx))

		webhookRows := endpointRows(t, subCtx, installation.ID)
		assert.Assert(t, is.Len(webhookRows, 1))
		endpointID := lo.FromPtr(webhookRows[0].EndpointID)
		secret := webhookRows[0].SecretToken

		v3 := versionedRuntime(t, testint.BuilderV3())
		v3Version := v3.Registry().Version(versionedDefinitionID.ID())

		assessment, err := v3.RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		assert.NilError(t, err)
		assert.Assert(t, assessment.Connection.Healthy)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		assert.Equal(t, v3Version, reloaded.DefinitionVersion)
		assert.Equal(t, testint.UserInputV3.Name(), reloaded.UserInput.Layout)
		assertJSONEqual(t, `{"filter":"initial"}`, string(reloaded.UserInput.Data))
		assertJSONEqual(t, `{"filter":"initial"}`, string(reloaded.OperationConfig.For(testint.SyncOpV3.Name())))
		assert.Assert(t, is.Nil(reloaded.OperationConfig.For(testint.SyncOp.Name())))
		assert.DeepEqual(t, map[string]string{testint.SyncOpV3.Name(): "boom"}, reloaded.Health.UnhealthyOperations)

		rows, err := store.LoadAllCredentials(subCtx, installation)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(rows, 1))
		assertJSONEqual(t, `{"token":"initial-token","region":"`+installation.ID+`"}`, string(rows[testint.TokenV3.Connection().Credential.Name].Data))
		assert.Assert(t, is.Len(slotRowIDs(t, subCtx, installation.ID, testint.TokenV1.Connection().Credential.Name), 0))

		movedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, testint.SyncOpV3.Name())
		assert.NilError(t, err)
		assert.Assert(t, movedAt != nil)
		assert.Assert(t, movedAt.Equal(*finishedAt))

		webhookRows = endpointRows(t, subCtx, installation.ID)
		assert.Assert(t, is.Len(webhookRows, 1))
		assert.Equal(t, testint.WebhookV3.Name(), webhookRows[0].Name)
		assert.Equal(t, endpointID, lo.FromPtr(webhookRows[0].EndpointID))
		assert.Equal(t, secret, webhookRows[0].SecretToken)
	})

	t.Run("replacing removed after installations converged", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		v1 := versionedRuntime(t, testint.BuilderV1())
		installation := installOn(t, subCtx, v1, json.RawMessage(`{"filterExpr":"initial"}`), syncOpConfig("initial"), testint.TokenV1.Connection().Credential.Name, testint.TokenV1Set("converged-token"))

		retired, err := v1.Registry().Operation(versionedDefinitionID.ID(), testint.SyncOp.Name())
		assert.NilError(t, err)

		finished, err := operations.CreatePendingRun(subCtx, suite.Client.DB, installation, retired, enums.IntegrationRunTypeManual, nil)
		assert.NilError(t, err)
		assert.NilError(t, operations.CompleteRun(subCtx, suite.Client.DB, finished.ID, finished.StartedAt, operations.RunResult{}))

		assert.NilError(t, suite.Client.DB.Integration.UpdateOneID(installation.ID).
			SetHealth(models.IntegrationHealth{UnhealthyOperations: map[string]string{testint.SyncOp.Name(): "boom"}}).
			Exec(subCtx))

		_, err = versionedRuntime(t, testint.BuilderV2()).RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		assert.NilError(t, err)

		v3 := versionedRuntime(t, testint.BuilderV3())

		_, err = v3.RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		assert.NilError(t, err)

		converged := reloadIntegration(t, subCtx, installation.ID)
		assert.Equal(t, v3.Registry().Version(versionedDefinitionID.ID()), converged.DefinitionVersion)

		convergedRuns, err := suite.Client.DB.IntegrationRun.Query().
			Where(integrationrun.IntegrationIDEQ(installation.ID)).
			Select(integrationrun.FieldOperationName).
			Strings(subCtx)
		assert.NilError(t, err)
		assert.DeepEqual(t, []string{testint.SyncOpV4.Name()}, convergedRuns)

		convergedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, testint.SyncOpV4.Name())
		assert.NilError(t, err)
		assert.Assert(t, convergedAt != nil)

		convergedWebhooks := endpointRows(t, subCtx, installation.ID)
		assert.Assert(t, is.Len(convergedWebhooks, 1))

		v4 := versionedRuntime(t, testint.BuilderV4())
		v4Version := v4.Registry().Version(versionedDefinitionID.ID())
		assert.Assert(t, converged.DefinitionVersion != v4Version)

		assessment, err := v4.RunHealthAssessment(subCtx, converged)
		assert.NilError(t, err)
		assert.Assert(t, assessment.Connection.Healthy)
		assert.DeepEqual(t, []intruntime.OperationHealthResult{{Name: testint.SyncOpV4.Name(), Reason: "boom"}}, assessment.Operations)

		upgraded := reloadIntegration(t, subCtx, installation.ID)
		assert.Equal(t, v4Version, upgraded.DefinitionVersion)
		assert.Equal(t, converged.Status, upgraded.Status)
		assert.Equal(t, "", upgraded.Health.UnhealthyReason)
		assert.Equal(t, converged.UserInput.Layout, upgraded.UserInput.Layout)
		assertJSONEqual(t, string(converged.UserInput.Data), string(upgraded.UserInput.Data))
		assertJSONEqual(t, string(converged.OperationConfig.For(testint.SyncOpV4.Name())), string(upgraded.OperationConfig.For(testint.SyncOpV4.Name())))
		assert.DeepEqual(t, converged.Health.UnhealthyOperations, upgraded.Health.UnhealthyOperations)

		rows, err := store.LoadAllCredentials(subCtx, installation)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(rows, 1))
		assertJSONEqual(t, `{"token":"converged-token","region":"`+installation.ID+`"}`, string(rows[testint.TokenV4.Connection().Credential.Name].Data))

		runs, err := suite.Client.DB.IntegrationRun.Query().
			Where(integrationrun.IntegrationIDEQ(installation.ID)).
			Select(integrationrun.FieldOperationName).
			Strings(subCtx)
		assert.NilError(t, err)
		assert.DeepEqual(t, convergedRuns, runs)

		upgradedAt, err := operations.LastSuccessfulRunAt(subCtx, suite.Client.DB, installation.ID, testint.SyncOpV4.Name())
		assert.NilError(t, err)
		assert.Assert(t, upgradedAt != nil)
		assert.Assert(t, upgradedAt.Equal(*convergedAt))

		webhookRows := endpointRows(t, subCtx, installation.ID)
		assert.Assert(t, is.Len(webhookRows, 1))
		assert.Equal(t, convergedWebhooks[0].ID, webhookRows[0].ID)
		assert.Equal(t, testint.WebhookV4.Name(), webhookRows[0].Name)
		assert.Equal(t, lo.FromPtr(convergedWebhooks[0].EndpointID), lo.FromPtr(webhookRows[0].EndpointID))
		assert.Equal(t, convergedWebhooks[0].SecretToken, webhookRows[0].SecretToken)
	})

	t.Run("replacing removed before an installation upgraded strands it", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		v1 := versionedRuntime(t, testint.BuilderV1())
		installation := installOn(t, subCtx, v1, json.RawMessage(`{"filterExpr":"initial"}`), nil, testint.TokenV1.Connection().Credential.Name, testint.TokenV1Set("stranded-token"))
		assert.Equal(t, v1.Registry().Version(versionedDefinitionID.ID()), installation.DefinitionVersion)

		installedWebhooks := endpointRows(t, subCtx, installation.ID)
		assert.Assert(t, is.Len(installedWebhooks, 1))

		v4 := versionedRuntime(t, testint.BuilderV4())

		_, err := v4.RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		assert.ErrorIs(t, err, intruntime.ErrInstallationUpgradeFailed)
		assert.ErrorIs(t, err, intruntime.ErrUserInputInvalid)

		_, unhealthy := integrationtypes.UnhealthyFrom(err)
		assert.Assert(t, unhealthy)

		stranded := reloadIntegration(t, subCtx, installation.ID)
		assert.Equal(t, installation.DefinitionVersion, stranded.DefinitionVersion)
		assert.Equal(t, enums.IntegrationStatusErrored, stranded.Status)
		assert.Assert(t, is.Contains(stranded.Health.UnhealthyReason, intruntime.ErrInstallationUpgradeFailed.Error()))
		assert.Assert(t, is.Contains(stranded.Health.UnhealthyReason, intruntime.ErrUserInputInvalid.Error()))
		assert.Equal(t, testint.UserInputV1.Name(), stranded.UserInput.Layout)
		assertJSONEqual(t, `{"filterExpr":"initial"}`, string(stranded.UserInput.Data))

		rows, err := store.LoadAllCredentials(subCtx, installation)
		assert.NilError(t, err)
		assert.Assert(t, is.Len(rows, 1))
		assertJSONEqual(t, `{"accessToken":"stranded-token"}`, string(rows[testint.TokenV1.Connection().Credential.Name].Data))
		assert.Assert(t, is.Len(slotRowIDs(t, subCtx, installation.ID, testint.TokenV4.Connection().Credential.Name), 0))

		webhookRows := endpointRows(t, subCtx, installation.ID)
		assert.Assert(t, is.Len(webhookRows, 1))
		assert.Equal(t, installedWebhooks[0].ID, webhookRows[0].ID)
		assert.Equal(t, testint.WebhookV1V2.Name(), webhookRows[0].Name)

		_, err = v4.RunHealthAssessment(subCtx, stranded)
		assert.ErrorIs(t, err, intruntime.ErrUserInputInvalid)
		assert.Equal(t, installation.DefinitionVersion, reloadIntegration(t, subCtx, installation.ID).DefinitionVersion)
	})

	t.Run("an installation with no recorded version is upgraded and stamped on first use", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		v3 := versionedRuntime(t, testint.BuilderV3())
		def, ok := v3.Registry().Definition(versionedDefinitionID.ID())
		assert.Assert(t, ok)

		state, err := def.WithProviderState(integrationtypes.IntegrationProviderState{}, integrationtypes.DefinitionProviderState{CredentialRef: testint.TokenV3.Connection().Credential.Name})
		assert.NilError(t, err)

		installation, err := suite.Client.DB.Integration.Create().
			SetName("No Version").
			SetKind(versionedDefinitionID.ID()).
			SetDefinitionID(versionedDefinitionID.ID()).
			SetDefinitionVersion("").
			SetProviderState(state).
			SetStatus(enums.IntegrationStatusConnected).
			Save(subCtx)
		assert.NilError(t, err)
		assert.Equal(t, "", installation.DefinitionVersion)

		assert.NilError(t, store.SaveCredential(subCtx, installation, testint.TokenV3.Connection().Credential.Name, integrationtypes.CredentialSet{Data: json.RawMessage(`{"token":"incomplete-token"}`)}))

		v3Version := v3.Registry().Version(versionedDefinitionID.ID())

		_, err = v3.RunHealthAssessment(subCtx, installation)
		assert.NilError(t, err)

		reloaded := reloadIntegration(t, subCtx, installation.ID)
		assert.Equal(t, v3Version, reloaded.DefinitionVersion)

		rows, err := store.LoadAllCredentials(subCtx, installation)
		assert.NilError(t, err)
		assertJSONEqual(t, `{"token":"incomplete-token","region":"`+installation.ID+`"}`, string(rows[testint.TokenV3.Connection().Credential.Name].Data))
	})

	t.Run("installation metadata is refreshed by the upgrade", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		v1 := versionedRuntime(t, testint.BuilderV1())
		installation := installOn(t, subCtx, v1, json.RawMessage(`{"filterExpr":"x"}`), nil, testint.TokenV1.Connection().Credential.Name, testint.TokenV1Set("meta-token"))

		before := reloadIntegration(t, subCtx, installation.ID).InstallationMetadata
		assert.Equal(t, installation.ID, before.Display.ExternalID)

		var beforeAttrs struct {
			Region string `json:"region"`
		}
		assert.NilError(t, json.Unmarshal(before.Attributes, &beforeAttrs))
		assert.Equal(t, "", beforeAttrs.Region)

		v3 := versionedRuntime(t, testint.BuilderV3())

		_, err := v3.RunHealthAssessment(subCtx, reloadIntegration(t, subCtx, installation.ID))
		assert.NilError(t, err)

		after := reloadIntegration(t, subCtx, installation.ID).InstallationMetadata
		assert.Assert(t, !reflect.DeepEqual(before, after))
		assert.Equal(t, installation.ID, after.Display.ExternalID)

		var attrs struct {
			Region string `json:"region"`
		}
		assert.NilError(t, json.Unmarshal(after.Attributes, &attrs))
		assert.Equal(t, installation.ID, attrs.Region)
	})

	t.Run("an operation executed with no explicit config resolves it from the installation's stored operation input", func(t *testing.T) {
		subOrg := suite.UserBuilder(context.Background(), t)
		subCtx := th.SetInternalContext(subOrg.UserCtx, suite.Client.DB)

		v3 := versionedRuntime(t, testint.BuilderV3())
		operationConfig := map[string]json.RawMessage{testint.SyncOpV3.Name(): json.RawMessage(`{"filter":"only-mine"}`)}
		installation := installOn(t, subCtx, v3, json.RawMessage(`{"filter":"global"}`), operationConfig, testint.TokenV3.Connection().Credential.Name, testint.TokenV3Set("token", "region-1"))
		assertJSONEqual(t, `{"filter":"only-mine"}`, string(installation.OperationConfig.For(testint.SyncOpV3.Name())))

		syncOp, err := v3.Registry().Operation(versionedDefinitionID.ID(), testint.SyncOpV3.Name())
		assert.NilError(t, err)

		_, err = v3.ExecuteOperation(subCtx, installation, syncOp, nil)
		assert.NilError(t, err)

		select {
		case raw := <-testint.SyncConfigV3:
			var cfg struct {
				Filter string `json:"filter"`
			}
			assert.NilError(t, json.Unmarshal(raw, &cfg))
			assert.Assert(t, cfg.Filter != "")
			assert.Equal(t, "only-mine", cfg.Filter)
		default:
			t.Fatal("expected the v3 reconcile handler to record its resolved config")
		}
	})
}
