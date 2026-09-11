//go:build test

package eventstest_test

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	th "github.com/theopenlane/core/v2/internal/graphapi/testharness"

	"github.com/samber/lo"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	openapi "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/finding"
	"github.com/theopenlane/core/v2/internal/ent/generated/integration"
	"github.com/theopenlane/core/v2/internal/ent/generated/privacy"
	"github.com/theopenlane/core/v2/internal/graphapi"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	intregistry "github.com/theopenlane/core/v2/internal/integrations/registry"
	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

const findingIngestTestOperation = "finding.ingest"

// findingIngestTestDefinition builds a minimal finding ingest definition whose mapping passes provider payloads through unchanged
func findingIngestTestDefinition(defID string) integrationtypes.Definition {
	passthrough := integrationtypes.MappingOverride{MapExpr: "payload"}

	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{
			ID:          defID,
			DisplayName: "Finding Ingest Test",
			Active:      true,
		},
		Operations: []integrationtypes.OperationRegistration{
			{
				Name:  findingIngestTestOperation,
				Topic: gala.TopicName("integration." + defID + "." + findingIngestTestOperation),
				IngestHandle: func(context.Context, integrationtypes.OperationRequest) ([]integrationtypes.IngestPayloadSet, error) {
					return nil, nil
				},
				Ingest: []integrationtypes.IngestContract{{Schema: entityops.SchemaFinding.Name}},
			},
		},
		Mappings: []integrationtypes.MappingRegistration{{Schema: entityops.SchemaFinding.Name, Spec: passthrough}},
	}
}

// ingestFindingPayloads pushes finding payloads through the synchronous catalog ingest path and returns the record-level result
func ingestFindingPayloads(ctx context.Context, t *testing.T, installation *ent.Integration, payloads ...string) operations.IngestResult {
	t.Helper()

	def := findingIngestTestDefinition(installation.DefinitionID)
	reg := intregistry.New()
	th.RequireNoError(t, reg.Register(def))

	envelopes := lo.Map(payloads, func(p string, _ int) integrationtypes.MappingEnvelope {
		return integrationtypes.MappingEnvelope{Payload: json.RawMessage(p)}
	})

	result, err := operations.ProcessPayloadSets(ctx, operations.IngestContext{
		Registry:    reg,
		DB:          suite.Client.DB,
		Integration: installation,
	}, findingIngestTestOperation, def.Operations[0].Ingest, def.Operations[0].Policy, []integrationtypes.IngestPayloadSet{
		{Schema: entityops.SchemaFinding.Name, Envelopes: envelopes},
	}, operations.IngestOptions{})
	th.RequireNoError(t, err)

	return result
}

// findingByExternalID loads one ingested finding by its lookup key
func findingByExternalID(ctx context.Context, t *testing.T, externalID string) *ent.Finding {
	t.Helper()

	f, err := suite.Client.DB.Finding.Query().Where(finding.ExternalID(externalID)).Only(ctx)
	th.RequireNoError(t, err)

	return f
}

// TestFindingVolatileOnlyReingestNoop verifies a Volatile-only field difference never writes on its own but rides along on a material change
func TestFindingVolatileOnlyReingestNoop(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	var findingCreates, findingUpdates atomic.Int64

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaFinding,
			Operations: []string{entityops.OpCreate, entityops.OpUpdateOne},
			Handle: func(_ entityops.Invocation, payload entityops.MutationPayload) error {
				if payload.Operation == entityops.OpCreate {
					findingCreates.Add(1)
				} else {
					findingUpdates.Add(1)
				}

				return nil
			},
		},
	})
	assert.NilError(t, err)
	defer setup.Teardown()

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Finding Volatile Test").
		SetKind("findvolatiletest").
		SetDefinitionID("def_findvolatiletest").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findvolatiletest"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		findings, err := suite.Client.DB.Finding.Query().Where(finding.ExternalIDHasPrefix("find-vol-")).All(ctx)
		th.RequireNoError(t, err)

		if len(findings) > 0 {
			(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: lo.Map(findings, func(f *ent.Finding, _ int) string { return f.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	initialEventTime := "2024-01-01T00:00:00Z"
	create := `{"external_id":"find-vol-1","display_name":"Volatile Finding","description":"initial desc","event_time":"` + initialEventTime + `"}`

	result := ingestFindingPayloads(ctx, t, installation, create)
	waitForGala(t, setup.Runtime)

	assert.Check(t, is.Equal(1, result.Persisted))
	assert.Check(t, is.Equal(1, result.Changed), "a created row must count as changed")
	assert.Check(t, is.Equal(int64(1), findingCreates.Load()))
	assert.Check(t, is.Equal(int64(0), findingUpdates.Load()))

	created := findingByExternalID(ctx, t, "find-vol-1")
	assert.Assert(t, created.EventTime != nil)

	volatileOnly := `{"external_id":"find-vol-1","display_name":"Volatile Finding","description":"initial desc","event_time":"2024-06-01T00:00:00Z"}`

	result = ingestFindingPayloads(ctx, t, installation, volatileOnly)
	waitForGala(t, setup.Runtime)

	assert.Check(t, is.Equal(1, result.Persisted))
	assert.Check(t, is.Equal(0, result.Changed), "a volatile-only payload must not count as changed")
	assert.Check(t, is.Equal(int64(1), findingCreates.Load()))
	assert.Check(t, is.Equal(int64(0), findingUpdates.Load()), "a volatile-only payload must not emit an update mutation event")

	afterVolatileOnly := findingByExternalID(ctx, t, "find-vol-1")
	assert.Check(t, afterVolatileOnly.UpdatedAt.Equal(created.UpdatedAt), "a volatile-only payload must not rewrite the row")
	assert.Check(t, time.Time(*afterVolatileOnly.EventTime).Equal(time.Time(*created.EventTime)), "an unwritten volatile field must keep its stored value")

	materialAndVolatile := `{"external_id":"find-vol-1","display_name":"Volatile Finding","description":"changed desc","event_time":"2024-06-01T00:00:00Z"}`

	result = ingestFindingPayloads(ctx, t, installation, materialAndVolatile)
	waitForGala(t, setup.Runtime)

	assert.Check(t, is.Equal(1, result.Persisted))
	assert.Check(t, is.Equal(1, result.Changed), "a material change must count as changed")
	assert.Check(t, is.Equal(int64(1), findingUpdates.Load()), "a material change must emit an update mutation event")

	after := findingByExternalID(ctx, t, "find-vol-1")
	assert.Check(t, is.Equal("changed desc", after.Description))
	assert.Check(t, !time.Time(*after.EventTime).Equal(time.Time(*created.EventTime)), "the volatile field must ride along once a material field changed")
	assert.Check(t, time.Time(*after.EventTime).Equal(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)))
}

// TestFindingForeignDefinitionReadOnly verifies a Finding claimed by one definition is read-only for a payload from a different definition
func TestFindingForeignDefinitionReadOnly(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	installationA, err := suite.Client.DB.Integration.Create().
		SetName("Finding Foreign Def A").
		SetKind("findforeigna").
		SetDefinitionID("def_findforeign_a").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findforeigna"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	installationB, err := suite.Client.DB.Integration.Create().
		SetName("Finding Foreign Def B").
		SetKind("findforeignb").
		SetDefinitionID("def_findforeign_b").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findforeignb"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		findings, err := suite.Client.DB.Finding.Query().Where(finding.ExternalIDHasPrefix("find-foreign-")).All(ctx)
		th.RequireNoError(t, err)

		if len(findings) > 0 {
			(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: lo.Map(findings, func(f *ent.Finding, _ int) string { return f.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationA.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationB.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	resultA := ingestFindingPayloads(ctx, t, installationA, `{"external_id":"find-foreign-1","display_name":"Owned By A","description":"from A"}`)
	assert.Check(t, is.Equal(1, resultA.Changed), "the claiming definition's create must count as changed")

	before := findingByExternalID(ctx, t, "find-foreign-1")

	linkedToA, err := suite.Client.DB.Finding.Query().Where(finding.ID(before.ID), finding.HasIntegrationsWith(integration.ID(installationA.ID))).Exist(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, linkedToA, "the claiming definition's create must add its integration edge")

	resultB := ingestFindingPayloads(ctx, t, installationB, `{"external_id":"find-foreign-1","display_name":"Owned By B","description":"from B"}`)
	assert.Check(t, is.Equal(1, resultB.Skipped), "a foreign-definition payload must count as skipped, not persisted")
	assert.Check(t, is.Equal(0, resultB.Changed), "a foreign-definition payload must be read-only")

	after := findingByExternalID(ctx, t, "find-foreign-1")
	assert.Check(t, is.Equal(before.Description, after.Description), "a foreign-definition payload must not change fields")
	assert.Check(t, is.Equal(before.DisplayName, after.DisplayName))
	assert.Check(t, after.UpdatedAt.Equal(before.UpdatedAt))

	linkedToB, err := suite.Client.DB.Finding.Query().Where(finding.ID(after.ID), finding.HasIntegrationsWith(integration.ID(installationB.ID))).Exist(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, !linkedToB, "a foreign-definition ingest must not add its integration edge")
}

// TestFindingClaimUnclaimedRowTakenOverInOneWrite verifies a finding ingest payload takes over a row
// that carries no recorded source definition, in the same write that applies its other field changes
func TestFindingClaimUnclaimedRowTakenOverInOneWrite(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Finding Claim Unclaimed Test").
		SetKind("findclaimunclaimed").
		SetDefinitionID("def_findclaimunclaimed").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findclaimunclaimed"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	unclaimed, err := suite.Client.DB.Finding.Create().
		SetExternalID("find-claim-unclaimed-1").
		SetDisplayName("Pre-existing Finding").
		Save(privacy.DecisionContext(ctx, privacy.Allow))
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, ID: unclaimed.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	result := ingestFindingPayloads(ctx, t, installation, `{"external_id":"find-claim-unclaimed-1","display_name":"Claimed Finding","description":"claimed"}`)
	assert.Check(t, is.Equal(0, result.Failed))
	assert.Check(t, is.Equal(1, result.Changed), "the display name change and the claim land in the same write")

	after := findingByExternalID(ctx, t, "find-claim-unclaimed-1")
	assert.Check(t, is.Equal(unclaimed.ID, after.ID), "the takeover must update the existing row, not create a new one")
	assert.Check(t, is.Equal("Claimed Finding", after.DisplayName))
	assert.Check(t, is.Equal(installation.ID, after.ManagedBy))

	linked, err := after.QueryIntegrations().Where(integration.ID(installation.ID)).Exist(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, linked, "the claim must add the claiming installation's integration edge")
}

// TestFindingClaimActiveOtherInstallationKeepsPointers verifies a finding managed by another
// still-active installation of the same definition keeps its ownership pointer while its other
// fields still diff and update normally
func TestFindingClaimActiveOtherInstallationKeepsPointers(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const sharedDefinitionID = "def_findclaimactive"

	installationA, err := suite.Client.DB.Integration.Create().
		SetName("Finding Claim Active A").
		SetKind("findclaimactivea").
		SetDefinitionID(sharedDefinitionID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findclaimactivea"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	installationB, err := suite.Client.DB.Integration.Create().
		SetName("Finding Claim Active B").
		SetKind("findclaimactiveb").
		SetDefinitionID(sharedDefinitionID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findclaimactiveb"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		findings, err := suite.Client.DB.Finding.Query().Where(finding.ExternalIDHasPrefix("find-claim-active-")).All(ctx)
		th.RequireNoError(t, err)

		if len(findings) > 0 {
			(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: lo.Map(findings, func(f *ent.Finding, _ int) string { return f.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationA.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationB.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	resultA := ingestFindingPayloads(ctx, t, installationA, `{"external_id":"find-claim-active-1","display_name":"Owned By A","description":"from A"}`)
	assert.Check(t, is.Equal(1, resultA.Changed))

	resultB := ingestFindingPayloads(ctx, t, installationB, `{"external_id":"find-claim-active-1","display_name":"Owned By A","description":"changed by B"}`)
	assert.Check(t, is.Equal(0, resultB.Failed))
	assert.Check(t, is.Equal(1, resultB.Changed), "the description change must still apply")

	after := findingByExternalID(ctx, t, "find-claim-active-1")
	assert.Check(t, is.Equal(installationA.ID, after.ManagedBy), "an active other installation must keep managed_by")
	assert.Check(t, is.Equal("changed by B", after.Description))

	linkedToB, err := after.QueryIntegrations().Where(integration.ID(installationB.ID)).Exist(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, !linkedToB, "an active other installation's claim must not add the non-owning installation's integration edge")
}

// TestFindingClaimGoneOtherInstallationRepoints verifies a finding whose managing installation no
// longer exists is repointed to the new installation in the same update as any other field change
func TestFindingClaimGoneOtherInstallationRepoints(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	const sharedDefinitionID = "def_findclaimgone"

	installationA, err := suite.Client.DB.Integration.Create().
		SetName("Finding Claim Gone A").
		SetKind("findclaimgonea").
		SetDefinitionID(sharedDefinitionID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findclaimgonea"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	installationB, err := suite.Client.DB.Integration.Create().
		SetName("Finding Claim Gone B").
		SetKind("findclaimgoneb").
		SetDefinitionID(sharedDefinitionID).
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findclaimgoneb"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	t.Cleanup(func() {
		findings, err := suite.Client.DB.Finding.Query().Where(finding.ExternalIDHasPrefix("find-claim-gone-")).All(ctx)
		th.RequireNoError(t, err)

		if len(findings) > 0 {
			(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: lo.Map(findings, func(f *ent.Finding, _ int) string { return f.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installationB.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	resultA := ingestFindingPayloads(ctx, t, installationA, `{"external_id":"find-claim-gone-1","display_name":"Owned By A","description":"from A"}`)
	assert.Check(t, is.Equal(1, resultA.Changed))

	th.RequireNoError(t, suite.Client.DB.Integration.DeleteOneID(installationA.ID).Exec(ctx))

	resultB := ingestFindingPayloads(ctx, t, installationB, `{"external_id":"find-claim-gone-1","display_name":"Adopted By B","description":"from A"}`)
	assert.Check(t, is.Equal(0, resultB.Failed))
	assert.Check(t, is.Equal(1, resultB.Changed), "the repoint and the display name change land in the same write")

	after := findingByExternalID(ctx, t, "find-claim-gone-1")
	assert.Check(t, is.Equal(installationB.ID, after.ManagedBy), "a gone manager must be repointed to the adopting installation")
	assert.Check(t, is.Equal("Adopted By B", after.DisplayName))

	linkedToB, err := after.QueryIntegrations().Where(integration.ID(installationB.ID)).Exist(ctx)
	th.RequireNoError(t, err)
	assert.Check(t, linkedToB, "repointing must add the adopting installation's integration edge")
}

// TestFindingStatusCasingFoldUnchanged verifies a provider casing that folds to the stored canonical value is unchanged
func TestFindingStatusCasingFoldUnchanged(t *testing.T) {
	ctx := th.SetContext(th.SharedTestUser1.UserCtx, suite.Client.DB)

	var findingUpdates atomic.Int64

	setup, err := graphapi.SetupListenerRuntime(suite.GalaRuntime, []gala.Registration{
		entityops.MutationListener{
			Schema:     entityops.SchemaFinding,
			Operations: []string{entityops.OpUpdateOne},
			Handle: func(_ entityops.Invocation, _ entityops.MutationPayload) error {
				findingUpdates.Add(1)

				return nil
			},
		},
	})
	assert.NilError(t, err)
	defer setup.Teardown()

	installation, err := suite.Client.DB.Integration.Create().
		SetName("Finding Casing Test").
		SetKind("findcasing").
		SetDefinitionID("def_findcasing").
		SetInstallationMetadata(openapi.IntegrationInstallationMetadata{Display: openapi.IntegrationInstallationIdentity{ExternalID: "tenant-findcasing"}}).
		Save(ctx)
	th.RequireNoError(t, err)

	// the test database carries no seeded system enums (db/seed/07_seed_status_exposure.sql is
	// production-only), so the custom-enum hook's canonical lookup needs its own system-owned row
	// to fold "NEW" against instead of auto-creating one from the raw provider casing
	statusEnum := (&th.CustomTypeEnumBuilder{
		Client:      suite.Client,
		Name:        "New",
		ObjectType:  "finding",
		Field:       "status",
		Description: "Newly identified by an external integration.",
		Color:       "#0B6623",
	}).MustNew(th.SharedSystemAdminUser.UserCtx, t)

	t.Cleanup(func() {
		(&th.Cleanup[*ent.CustomTypeEnumDeleteOne]{Client: suite.Client.DB.CustomTypeEnum, ID: statusEnum.ID}).MustDelete(th.SharedSystemAdminUser.UserCtx, t)
	})

	t.Cleanup(func() {
		findings, err := suite.Client.DB.Finding.Query().Where(finding.ExternalIDHasPrefix("casing-f-")).All(ctx)
		th.RequireNoError(t, err)

		if len(findings) > 0 {
			(&th.Cleanup[*ent.FindingDeleteOne]{Client: suite.Client.DB.Finding, IDs: lo.Map(findings, func(f *ent.Finding, _ int) string { return f.ID })}).MustDelete(th.SharedTestUser1.UserCtx, t)
		}

		(&th.Cleanup[*ent.IntegrationDeleteOne]{Client: suite.Client.DB.Integration, ID: installation.ID}).MustDelete(th.SharedTestUser1.UserCtx, t)
	})

	create := `{"external_id":"casing-f-1","display_name":"Casing Finding","finding_status_name":"NEW"}`

	result := ingestFindingPayloads(ctx, t, installation, create)
	assert.Check(t, is.Equal(1, result.Changed))

	before := findingByExternalID(ctx, t, "casing-f-1")
	assert.Check(t, is.Equal("New", before.FindingStatusName), "the custom-enum hook normalizes create-time casing to the seeded canonical form")

	result = ingestFindingPayloads(ctx, t, installation, create)
	waitForGala(t, setup.Runtime)

	assert.Check(t, is.Equal(0, result.Changed), "provider casing that folds to the stored canonical value must not count as changed")
	assert.Check(t, is.Equal(int64(0), findingUpdates.Load()), "a casing-only payload must not emit an update mutation event")

	after := findingByExternalID(ctx, t, "casing-f-1")
	assert.Check(t, after.UpdatedAt.Equal(before.UpdatedAt), "a casing-only payload must not rewrite the row")
}
