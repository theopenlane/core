//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/directorygroup"
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

	// membership payloads carry the member account and group by external id, so the mapping
	// declares link rules resolving those external ids to the actual DirectoryAccount and
	// DirectoryGroup rows, mirroring the link rules real directory mappings declare
	membershipLinks := integrationtypes.MappingOverride{
		MapExpr: "payload",
		Links: []integrationtypes.LinkRule{
			{TargetSchema: entityops.SchemaDirectoryAccount.Name, TargetField: "external_id", SourceField: "directory_account_id"},
			{TargetSchema: entityops.SchemaDirectoryGroup.Name, TargetField: "external_id", SourceField: "directory_group_id"},
		},
	}

	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          defID,
			DisplayName: "Directory Sync Test",
			Active:      true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:   directorySyncTestOperation,
				Topic:  gala.TopicName("integration." + defID + "." + directorySyncTestOperation),
				Policy: integrationtypes.ExecutionPolicy{Snapshot: true},
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
			{Schema: entityops.SchemaDirectoryMembership.Name, Spec: membershipLinks},
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
	}, directorySyncTestOperation, def.Operations[0].Ingest, def.Operations[0].Policy, []integrationtypes.IngestPayloadSet{
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

// TestDirectorySyncIngestUnchangedFieldGate verifies the directory sync ingest path creates,
// leaves unchanged, and updates rows correctly across the account and group schemas; renamed from
// TestDirectorySyncIngestProfileHashing now that change detection is a direct field comparison
// (pruneIngestFields) rather than a stored profile_hash column, which no longer exists
func TestDirectorySyncIngestUnchangedFieldGate(t *testing.T) {
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
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-dirsynctest"}}).
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

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: integration.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	accountPayload := `{"external_id":"dirhash-acct-1","canonical_email":"dirhash1@example.com","display_name":"Hash User","phone_number":"800-867-5309","profile":{"id":"dirhash-acct-1","displayName":"Hash User","rev":1}}`

	t.Run("account ingest stores the profile hash and emits a create", func(t *testing.T) {
		result := ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryAccount.Name, accountPayload)
		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(1, result.Persisted))
		assert.Check(t, is.Equal(0, result.Failed))
		assert.Check(t, is.Equal(1, result.Changed), "a created row must count as changed")
		assert.Check(t, is.Equal(int64(1), accountCreates.Load()))
		assert.Check(t, is.Equal(int64(0), accountUpdates.Load()))

		da := directoryAccountByExternalID(ctx, t, "dirhash-acct-1")
		assert.Check(t, is.Equal("800-867-5309", lo.FromPtr(da.PhoneNumber)))
		assert.Check(t, da.FirstSeenAt != nil)
		assert.Check(t, da.LastSeenAt == nil, "create stamps first seen only; last seen arrives with the next confirming sync")
	})

	t.Run("unchanged re-ingest advances last seen without emitting events", func(t *testing.T) {
		before := directoryAccountByExternalID(ctx, t, "dirhash-acct-1")

		result := ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryAccount.Name, accountPayload)
		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(1, result.Persisted))
		assert.Check(t, is.Equal(0, result.Changed), "an unchanged payload must not count as changed")
		assert.Check(t, is.Equal(int64(1), accountCreates.Load()))
		assert.Check(t, is.Equal(int64(0), accountUpdates.Load()), "an unchanged payload must not emit an update mutation event")

		after := directoryAccountByExternalID(ctx, t, "dirhash-acct-1")
		assert.Check(t, is.Equal(before.DisplayName, after.DisplayName))
		assert.Check(t, is.DeepEqual(before.Profile, after.Profile), "an unchanged payload must not rewrite the stored profile")
		assert.Check(t, after.LastSeenAt != nil, "the vetoed bookkeeping write must still confirm the sighting")
	})

	t.Run("changed payload updates the row and emits an update", func(t *testing.T) {
		changed := `{"external_id":"dirhash-acct-1","canonical_email":"dirhash1@example.com","display_name":"Hash User Two","phone_number":"800-867-5309","profile":{"id":"dirhash-acct-1","displayName":"Hash User Two","rev":2}}`

		result := ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryAccount.Name, changed)
		waitForGala(t, setup.Runtime)

		assert.Check(t, is.Equal(1, result.Persisted))
		assert.Check(t, is.Equal(1, result.Changed), "a changed payload must count as changed")
		assert.Check(t, is.Equal(int64(1), accountUpdates.Load()), "a changed payload must emit an update mutation event")

		after := directoryAccountByExternalID(ctx, t, "dirhash-acct-1")
		assert.Check(t, is.Equal("Hash User Two", after.DisplayName))
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

		result = ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryGroup.Name, groupPayload)
		assert.Check(t, is.Equal(1, result.Persisted))

		unchanged := directoryGroupByExternalID(ctx, t, "dirhash-grp-1")
		assert.Check(t, is.Equal(before.DisplayName, unchanged.DisplayName))
		assert.Check(t, is.DeepEqual(before.Profile, unchanged.Profile), "an unchanged group must not rewrite the stored profile")
		assert.Check(t, unchanged.LastSeenAt != nil, "the vetoed bookkeeping write must still confirm the sighting")

		changed := `{"external_id":"dirhash-grp-1","display_name":"Hash Group Two","profile":{"id":"dirhash-grp-1","rev":2}}`
		result = ingestDirectoryPayloads(ctx, t, integration, entityops.SchemaDirectoryGroup.Name, changed)
		assert.Check(t, is.Equal(1, result.Persisted))

		after := directoryGroupByExternalID(ctx, t, "dirhash-grp-1")
		assert.Check(t, is.Equal("Hash Group Two", after.DisplayName))
	})
}

// TestDirectoryAccountReinstallRelinkNoUpdate verifies that a reinstall under a new installation of
// the same source definition and instance is recognized as unchanged: integration_id and managed_by
// are Volatile on the generated DirectoryAccount descriptor, so a payload that differs only in those
// columns prunes to an empty change set and no per-row update or mutation event fires. Provenance
// repoints only ride along an actual write, so with nothing else to write, integration_id stays on
// the original installation until some other field changes.
func TestDirectoryAccountReinstallRelinkNoUpdate(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	var accountUpdates atomic.Int64

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaDirectoryAccount,
			Operations: []string{entityops.OpUpdateOne},
			Handle: func(_ entityops.Invocation, _ entityops.MutationPayload) error {
				accountUpdates.Add(1)

				return nil
			},
		},
	})
	assert.NilError(t, err)
	defer setup.Teardown()

	metadata := openapi.IntegrationInstallationMetadata{
		Display: openapi.IntegrationInstallationIdentity{ExternalID: "reinstall-tenant-1"},
	}

	installationA, err := suite.Client.DB.Integration.Create().
		SetName("Reinstall Test A").
		SetKind("reinstalla").
		SetDefinitionID("def_reinstall_a").
		SetInstallationMetadata(metadata).
		Save(ctx)
	th.RequireNoError(t, err)

	installationB, err := suite.Client.DB.Integration.Create().
		SetName("Reinstall Test B").
		SetKind("reinstallb").
		SetDefinitionID("def_reinstall_a").
		SetInstallationMetadata(metadata).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		accounts, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalIDHasPrefix("reinstall-")).All(ctx)
		th.RequireNoError(t, err)

		if len(accounts) > 0 {
			(&th.Cleanup[*ent.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, IDs: lo.Map(accounts, func(da *ent.DirectoryAccount, _ int) string { return da.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationA.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationB.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	accountPayload := `{"external_id":"reinstall-acct-1","canonical_email":"reinstall1@example.com","display_name":"Reinstall User"}`

	ingestDirectoryPayloads(ctx, t, installationA, entityops.SchemaDirectoryAccount.Name, accountPayload)
	waitForGala(t, setup.Runtime)

	before := directoryAccountByExternalID(ctx, t, "reinstall-acct-1")
	assert.Check(t, is.Equal(installationA.ID, before.IntegrationID))

	result := ingestDirectoryPayloads(ctx, t, installationB, entityops.SchemaDirectoryAccount.Name, accountPayload)
	waitForGala(t, setup.Runtime)

	assert.Check(t, is.Equal(0, result.Changed), "a reinstall adoption must not count as a material change")
	assert.Check(t, is.Equal(int64(0), accountUpdates.Load()), "a reinstall adoption must not emit a per-row update mutation event")

	after := directoryAccountByExternalID(ctx, t, "reinstall-acct-1")
	assert.Check(t, is.Equal(installationA.ID, after.IntegrationID), "integration_id stays on the original installation: provenance repoints only ride along an actual write, and no other field changed to justify one")
}

// TestDirectoryAccountProfileChurnUnchanged is regen-gated on defect 2 (profile/metadata = whole
// payload defeats Volatile): every directory mapping maps DirectoryAccountFields.Profile.Expr to
// the raw provider payload, profile is not Volatile, and profile_hash is derived from it, so a
// provider-only churn key nested inside profile (e.g. Okta/Google lastLoginTime, Tailscale
// LastSeen) rewrites the row every sync even though every mapped scalar field is unchanged. This
// test asserts the intended post-fix behavior — a churn-only profile key is treated as unchanged —
// and will fail until profile/profile_hash are marked Volatile or provider churn keys are stripped
// before hashing.
func TestDirectoryAccountProfileChurnUnchanged(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	var accountUpdates atomic.Int64

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaDirectoryAccount,
			Operations: []string{entityops.OpUpdateOne},
			Handle: func(_ entityops.Invocation, _ entityops.MutationPayload) error {
				accountUpdates.Add(1)

				return nil
			},
		},
	})
	assert.NilError(t, err)
	defer setup.Teardown()

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Profile Churn Test").
		SetKind("profilechurn").
		SetDefinitionID("def_profilechurn").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-profilechurn"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		accounts, err := suite.Client.DB.DirectoryAccount.Query().Where(directoryaccount.ExternalIDHasPrefix("churn-")).All(ctx)
		th.RequireNoError(t, err)

		if len(accounts) > 0 {
			(&th.Cleanup[*ent.DirectoryAccountDeleteOne]{Client: suite.Client.DB.DirectoryAccount, IDs: lo.Map(accounts, func(da *ent.DirectoryAccount, _ int) string { return da.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	initial := `{"external_id":"churn-acct-1","canonical_email":"churn1@example.com","display_name":"Churn User","profile":{"id":"churn-acct-1","displayName":"Churn User","lastLoginTime":"2024-01-01T00:00:00Z"}}`

	ingestDirectoryPayloads(ctx, t, installation, entityops.SchemaDirectoryAccount.Name, initial)

	before := directoryAccountByExternalID(ctx, t, "churn-acct-1")

	churned := `{"external_id":"churn-acct-1","canonical_email":"churn1@example.com","display_name":"Churn User","profile":{"id":"churn-acct-1","displayName":"Churn User","lastLoginTime":"2024-06-01T00:00:00Z"}}`

	result := ingestDirectoryPayloads(ctx, t, installation, entityops.SchemaDirectoryAccount.Name, churned)
	waitForGala(t, setup.Runtime)

	assert.Check(t, is.Equal(0, result.Changed), "a churn-only profile key must not count as changed")
	assert.Check(t, is.Equal(int64(0), accountUpdates.Load()), "a churn-only profile key must not emit an update mutation event")

	after := directoryAccountByExternalID(ctx, t, "churn-acct-1")
	assert.Check(t, is.DeepEqual(before.Profile, after.Profile), "a churn-only profile key must not be written to the stored profile")
	assert.Check(t, after.LastSeenAt != nil, "the vetoed bulk confirmation still advances last_seen_at by design")
}

// TestDirectoryMembershipStaleRunUnchangedNoop verifies a batched run whose id is older than the
// rows already in scope persists every unchanged record normally (the per-record run guard never
// fires for a byte-identical replay, since nothing changes for save to gate), but the scope-level
// staleness check vetoes snapshot removal for that run even though the payload appears to omit a
// membership
func TestDirectoryMembershipStaleRunUnchangedNoop(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const prefix = "stalerun"

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Stale Run Test").
		SetKind("stalerun").
		SetDefinitionID("def_dirsynctest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-stalerun"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	cleanupDirectoryPrefix(t, ctx, []string{installation.ID}, prefix)

	// runOlder is created first, so it carries a smaller ULID than runNewer; the replay below
	// ingests under runOlder against rows already stamped with runNewer, so the scope's run id is
	// greater than the replay's
	runOlder, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	runNewer, err := suite.Client.DB.IntegrationRun.Create().
		SetIntegrationID(installation.ID).
		SetOwnerID(installation.OwnerID).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.IntegrationRunDeleteOne]{Client: suite.Client.DB.IntegrationRun, IDs: []string{runOlder.ID, runNewer.ID}}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	base := newDirectorySnapshot(prefix)
	omitted := base.Memberships[0]

	seeded := ingestDirectorySnapshotFixture(ctx, t, installation, base, true, operations.IngestOptions{RunID: runNewer.ID})
	waitForGala(t, suite.GalaRuntime)
	assert.Check(t, is.Equal(0, seeded.Failed))

	staleAttempt := base.withoutMembership(omitted.DirectoryAccountID, omitted.DirectoryGroupID)

	result := ingestDirectorySnapshotFixture(ctx, t, installation, staleAttempt, true, operations.IngestOptions{RunID: runOlder.ID})
	waitForGala(t, suite.GalaRuntime)

	assert.Check(t, is.Equal(0, result.Failed))
	assert.Check(t, is.Equal(0, result.Removed), "a run older than the scope it targets must never perform snapshot removal")
	assert.Check(t, is.Equal(len(staleAttempt.Accounts)+len(staleAttempt.Groups)+len(staleAttempt.Memberships), result.Persisted), "every unchanged record must still persist (confirm) under an older run id, not be skipped")

	stillActive := directoryMembershipByExternalIDs(ctx, t, omitted.DirectoryAccountID, omitted.DirectoryGroupID)
	assert.Check(t, stillActive.RemovedAt == nil, "an older run must not mark an omitted membership removed")
	assert.Check(t, is.Equal(0, directoryRemovedMembershipCount(ctx, t, installation.ID)), "an older run must not remove anything")
}
