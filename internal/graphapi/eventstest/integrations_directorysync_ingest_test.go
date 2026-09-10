//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorysyncrun"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	intregistry "github.com/theopenlane/core/v2/internal/integrations/registry"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

const directorySyncTestOperation = "directory.sync"

// directorySyncTestDefinition builds a minimal directory ingest definition whose mappings pass
// provider payloads through unchanged, mirroring how real directory mappings carry the payload
// as the mapped document
func directorySyncTestDefinition(defID string) integrationtypes.Definition {
	passthrough := integrationtypes.MappingOverride{MapExpr: "payload"}

	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          defID,
			DisplayName: "Directory Sync Test",
			Active:      true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:  directorySyncTestOperation,
				Topic: gala.TopicName("integration." + defID + "." + directorySyncTestOperation),
				IngestHandle: func(context.Context, integrationtypes.OperationRequest) ([]integrationtypes.IngestPayloadSet, error) {
					return nil, nil
				},
				Ingest: []integrationtypes.IngestContract{
					{Schema: entityops.SchemaDirectoryAccount.Name},
					{Schema: entityops.SchemaDirectoryGroup.Name},
					{Schema: entityops.SchemaDirectoryMembership.Name},
				},
			},
		},
		Mappings: []integrationtypes.MappingRegistration{
			{Schema: entityops.SchemaDirectoryAccount.Name, Spec: passthrough},
			{Schema: entityops.SchemaDirectoryGroup.Name, Spec: passthrough},
			{Schema: entityops.SchemaDirectoryMembership.Name, Spec: passthrough},
		},
	}
}

// ingestDirectoryPayloads pushes payloads for one schema through the synchronous directory ingest
// path — mapping, generated preparation with sanitization, and hash-gated persistence — exactly as
// a reconcile cycle would, and returns the record-level result
func ingestDirectoryPayloads(ctx context.Context, t *testing.T, integration *ent.Integration, schema string, payloads ...string) operations.IngestResult {
	t.Helper()

	def := directorySyncTestDefinition(integration.DefinitionID)
	reg := intregistry.New()
	th.RequireNoError(t, reg.Register(def))

	envelopes := lo.Map(payloads, func(p string, _ int) integrationtypes.MappingEnvelope {
		return integrationtypes.MappingEnvelope{Payload: json.RawMessage(p)}
	})

	result, err := operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    reg,
		DB:          suite.Client.DB,
		Integration: integration,
	}, directorySyncTestOperation, def.Operations[0].Ingest, []integrationtypes.IngestPayloadSet{
		{Schema: schema, Envelopes: envelopes},
	}, operations.IngestOptions{})
	th.RequireNoError(t, err)

	return result
}

// directoryAccountByExternalID loads one ingested directory account by its external id
func directoryAccountByExternalID(ctx context.Context, t *testing.T, externalID string) *ent.DirectoryAccount {
	t.Helper()

	da, err := suite.Client.DB.DirectoryAccount.Query().
		Where(directoryaccount.ExternalID(externalID)).
		Only(ctx)
	th.RequireNoError(t, err)

	return da
}

// directoryGroupByExternalID loads one ingested directory group by its external id
func directoryGroupByExternalID(ctx context.Context, t *testing.T, externalID string) *ent.DirectoryGroup {
	t.Helper()

	dg, err := suite.Client.DB.DirectoryGroup.Query().
		Where(directorygroup.ExternalID(externalID)).
		Only(ctx)
	th.RequireNoError(t, err)

	return dg
}

// latestSyncRun loads the most recently started directory sync run for the integration
func latestSyncRun(ctx context.Context, t *testing.T, integrationID string) *ent.DirectorySyncRun {
	t.Helper()

	run, err := suite.Client.DB.DirectorySyncRun.Query().
		Where(directorysyncrun.IntegrationID(integrationID)).
		Order(directorysyncrun.ByStartedAt(entsql.OrderDesc())).
		First(ctx)
	th.RequireNoError(t, err)

	return run
}

func TestDirectorySyncIngestProfileHashing(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	// the counting listener creates mutation-topic interest, so the persist paths emit events for
	// it exactly as they would for identity resolution in production; the counters prove which
	// ingests emitted and which were suppressed by the unchanged-row gates
	var accountCreates, accountUpdates atomic.Int64

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaDirectoryAccount,
			Operations: []string{entityops.OpCreate, entityops.OpUpdateOne},
			Handle: func(_ entityops.Invocation, payload entityops.MutationPayload) error {
				if payload.Operation == entityops.OpCreate {
					accountCreates.Add(1)
				} else {
					accountUpdates.Add(1)
				}

				return nil
			},
		},
	})
	assert.NilError(t, err)
	defer setup.Teardown()

	integration, err := suite.Client.DB.Integration.Create().
		SetName("Directory Sync Hash Test").
		SetKind("dirsynctest").
		SetDefinitionID("def_dirsynctest").
		Save(ctx)
	th.RequireNoError(t, err)
	assert.Assert(t, integration.OwnerID != "", "seeded integration must be org-owned")

	t.Cleanup(func() {
		accounts, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalIDHasPrefix("dirhash-")).All(ctx)
		th.RequireNoError(t, err)

		if len(accounts) > 0 {
			(&th.Cleanup[*ent.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, IDs: lo.Map(accounts, func(da *ent.DirectoryAccount, _ int) string { return da.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		groups, err := suite.Client.DB.DirectoryGroup.Query().Where(directorygroup.ExternalIDHasPrefix("dirhash-")).All(ctx)
		th.RequireNoError(t, err)

		if len(groups) > 0 {
			(&th.Cleanup[*ent.DirectoryGroupDeleteOne]{Client: suite.Client.DB.DirectoryGroup, IDs: lo.Map(groups, func(dg *ent.DirectoryGroup, _ int) string { return dg.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		runs, err := suite.Client.DB.DirectorySyncRun.Query().Where(directorysyncrun.IntegrationID(integration.ID)).All(ctx)
		th.RequireNoError(t, err)

		if len(runs) > 0 {
			(&th.Cleanup[*ent.DirectorySyncRunDeleteOne]{Client: suite.Client.DB.DirectorySyncRun, IDs: lo.Map(runs, func(r *ent.DirectorySyncRun, _ int) string { return r.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	accountPayload := `{"external_id":"dirhash-acct-1","canonical_email":"dirhash1@example.com","display_name":"Hash User","phone_number":"800-867-5309","profile":{"id":"dirhash-acct-1","displayName":"Hash User","rev":1}}`

	t.Run("account ingest stores the profile hash and emits a create", func(t *testing.T) {
		result := ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryAccount.Name, accountPayload)
		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(1, result.Persisted))
		assert.Check(t, is.Equal(0, result.Failed))
		assert.Check(t, is.Equal(int64(1), accountCreates.Load()))
		assert.Check(t, is.Equal(int64(0), accountUpdates.Load()))

		da := directoryAccountByExternalID(ctx, t, "dirhash-acct-1")
		assert.Check(t, is.Len(da.ProfileHash, 64), "profile hash must be stored as a sha256 hex digest")
		assert.Check(t, is.Equal("800-867-5309", lo.FromPtr(da.PhoneNumber)))
		assert.Check(t, da.FirstSeenAt != nil)
		assert.Check(t, da.LastSeenAt == nil, "create stamps first seen only; last seen arrives with the next confirming sync")
	})

	t.Run("unchanged re-ingest advances last seen without emitting events", func(t *testing.T) {
		before := directoryAccountByExternalID(ctx, t, "dirhash-acct-1")

		result := ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryAccount.Name, accountPayload)
		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(1, result.Persisted))
		assert.Check(t, is.Equal(int64(1), accountCreates.Load()))
		assert.Check(t, is.Equal(int64(0), accountUpdates.Load()), "an unchanged payload must not emit an update mutation event")

		after := directoryAccountByExternalID(ctx, t, "dirhash-acct-1")
		assert.Check(t, is.Equal(before.ProfileHash, after.ProfileHash))
		assert.Check(t, is.Equal("Hash User", after.DisplayName))
		assert.Check(t, after.LastSeenAt != nil, "the vetoed bookkeeping write must still confirm the sighting")
	})

	t.Run("changed payload updates the row and emits an update", func(t *testing.T) {
		before := directoryAccountByExternalID(ctx, t, "dirhash-acct-1")

		changed := `{"external_id":"dirhash-acct-1","canonical_email":"dirhash1@example.com","display_name":"Hash User Two","phone_number":"800-867-5309","profile":{"id":"dirhash-acct-1","displayName":"Hash User Two","rev":2}}`

		result := ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryAccount.Name, changed)
		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(1, result.Persisted))
		assert.Check(t, is.Equal(int64(1), accountUpdates.Load()), "a changed payload must emit an update mutation event")

		after := directoryAccountByExternalID(ctx, t, "dirhash-acct-1")
		assert.Check(t, is.Equal("Hash User Two", after.DisplayName))
		assert.Check(t, before.ProfileHash != after.ProfileHash, "the stored hash must follow the changed payload")
	})

	t.Run("invalid optional fields are dropped instead of failing the record", func(t *testing.T) {
		payload := `{"external_id":"dirhash-acct-2","canonical_email":"dirhash2@example.com","phone_number":"not-a-phone","email_aliases":["not-an-email"],"avatar_remote_url":"::bad::","profile":{"id":"dirhash-acct-2"}}`

		result := ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryAccount.Name, payload)
		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(1, result.Persisted), "the record must persist despite invalid optional fields")
		assert.Check(t, is.Equal(0, result.Failed))

		da := directoryAccountByExternalID(ctx, t, "dirhash-acct-2")
		assert.Check(t, da.PhoneNumber == nil, "an unparsable phone number must be dropped")
		assert.Check(t, is.Len(da.EmailAliases, 0), "invalid email aliases must be dropped")
		assert.Check(t, da.AvatarRemoteURL == nil, "an invalid avatar URL must be dropped")
		assert.Check(t, is.Equal("dirhash2@example.com", lo.FromPtr(da.CanonicalEmail)))
	})

	groupPayload := `{"external_id":"dirhash-grp-1","display_name":"Hash Group","profile":{"id":"dirhash-grp-1","rev":1}}`

	t.Run("unchanged group re-ingest skips the update entirely", func(t *testing.T) {
		result := ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryGroup.Name, groupPayload)
		assert.Check(t, is.Equal(1, result.Persisted))

		before := directoryGroupByExternalID(ctx, t, "dirhash-grp-1")
		assert.Check(t, is.Len(before.ProfileHash, 64))

		result = ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryGroup.Name, groupPayload)
		assert.Check(t, is.Equal(1, result.Persisted))

		unchanged := directoryGroupByExternalID(ctx, t, "dirhash-grp-1")
		assert.Check(t, unchanged.UpdatedAt.Equal(before.UpdatedAt), "an unchanged group must not be rewritten")

		changed := `{"external_id":"dirhash-grp-1","display_name":"Hash Group Two","profile":{"id":"dirhash-grp-1","rev":2}}`
		result = ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryGroup.Name, changed)
		assert.Check(t, is.Equal(1, result.Persisted))

		after := directoryGroupByExternalID(ctx, t, "dirhash-grp-1")
		assert.Check(t, is.Equal("Hash Group Two", after.DisplayName))
		assert.Check(t, before.ProfileHash != after.ProfileHash)
	})

	t.Run("sync run records the processed count and no error when clean", func(t *testing.T) {
		ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryAccount.Name, accountPayload)

		run := latestSyncRun(ctx, t, integration.ID)
		assert.Check(t, is.Equal(enums.DirectorySyncRunStatusCompleted, run.Status))
		assert.Check(t, is.Equal(1, run.FullCount))
		assert.Check(t, run.Error == nil)
	})

	t.Run("sync run carries a failure summary when records are skipped", func(t *testing.T) {
		membership := `{"directory_account_id":"dirhash-missing-account","directory_group_id":"dirhash-grp-1"}`

		result := ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryMembership.Name, membership)
		assert.Check(t, is.Equal(1, result.Attempted))
		assert.Check(t, is.Equal(1, result.Failed))

		run := latestSyncRun(ctx, t, integration.ID)
		assert.Check(t, is.Equal(enums.DirectorySyncRunStatusCompleted, run.Status), "record-level failures must not fail the run")
		assert.Check(t, is.Equal(1, run.FullCount))
		assert.Assert(t, run.Error != nil, "a completed run with skipped records must carry a failure summary")
		assert.Check(t, is.Contains(*run.Error, "1 of 1 records failed"))
		assert.Check(t, is.Contains(*run.Error, "unresolved directory account"))
	})
}
