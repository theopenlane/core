package operations

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestWithDirectorySyncRunID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// default is empty
	if got := directorySyncRunIDFromContext(ctx); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}

	// set and retrieve
	ctx = withDirectorySyncRunID(ctx, "run-123")
	if got := directorySyncRunIDFromContext(ctx); got != "run-123" {
		t.Fatalf("expected %q, got %q", "run-123", got)
	}
}

func TestLookupIngestSchemaRegistration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		schema string
		want   bool
	}{
		{"asset schema exists", integrationgenerated.IntegrationMappingSchemaAsset, true},
		{"contact schema exists", integrationgenerated.IntegrationMappingSchemaContact, true},
		{"directory account exists", integrationgenerated.IntegrationMappingSchemaDirectoryAccount, true},
		{"directory group exists", integrationgenerated.IntegrationMappingSchemaDirectoryGroup, true},
		{"directory membership exists", integrationgenerated.IntegrationMappingSchemaDirectoryMembership, true},
		{"entity exists", integrationgenerated.IntegrationMappingSchemaEntity, true},
		{"finding exists", integrationgenerated.IntegrationMappingSchemaFinding, true},
		{"risk exists", integrationgenerated.IntegrationMappingSchemaRisk, true},
		{"vulnerability exists", integrationgenerated.IntegrationMappingSchemaVulnerability, true},
		{"unknown schema missing", "nonexistent", false},
		{"empty string missing", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, ok := lookupIngestSchemaRegistration(tc.schema)
			if ok != tc.want {
				t.Fatalf("lookupIngestSchemaRegistration(%q)=%v, want %v", tc.schema, ok, tc.want)
			}
		})
	}
}

func TestBuildIngestMetadata(t *testing.T) {
	t.Parallel()

	integration := &ent.Integration{
		ID:           "int-001",
		DefinitionID: "def-001",
	}

	t.Run("basic metadata without workflow", func(t *testing.T) {
		t.Parallel()

		record := mappedIngestRecord{
			Schema:  "asset",
			Variant: "server",
		}

		options := IngestOptions{
			Source:       integrationgenerated.IntegrationIngestSourceWebhook,
			RunID:        "run-001",
			Webhook:      "github",
			WebhookEvent: "push",
			DeliveryID:   "delivery-001",
		}

		got := buildIngestMetadata(integration, "health", record, options)

		if got.IntegrationID != "int-001" {
			t.Fatalf("IntegrationID=%q, want %q", got.IntegrationID, "int-001")
		}
		if got.DefinitionID != "def-001" {
			t.Fatalf("DefinitionID=%q, want %q", got.DefinitionID, "def-001")
		}
		if got.Operation != "health" {
			t.Fatalf("Operation=%q, want %q", got.Operation, "health")
		}
		if got.Variant != "server" {
			t.Fatalf("Variant=%q, want %q", got.Variant, "server")
		}
		if got.Source != integrationgenerated.IntegrationIngestSourceWebhook {
			t.Fatalf("Source=%q, want %q", got.Source, integrationgenerated.IntegrationIngestSourceWebhook)
		}
		if got.RunID != "run-001" {
			t.Fatalf("RunID=%q, want %q", got.RunID, "run-001")
		}
		if got.WorkflowInstanceID != "" {
			t.Fatalf("WorkflowInstanceID=%q, want empty", got.WorkflowInstanceID)
		}
	})

	t.Run("defaults source to direct", func(t *testing.T) {
		t.Parallel()

		got := buildIngestMetadata(integration, "sync", mappedIngestRecord{Schema: "asset"}, IngestOptions{})

		if got.Source != integrationgenerated.IntegrationIngestSourceDirect {
			t.Fatalf("Source=%q, want %q", got.Source, integrationgenerated.IntegrationIngestSourceDirect)
		}
	})

	t.Run("includes workflow metadata", func(t *testing.T) {
		t.Parallel()

		options := IngestOptions{
			WorkflowMeta: &types.WorkflowMeta{
				InstanceID:  "wf-001",
				ActionKey:   "action-key",
				ActionIndex: 3,
				ObjectID:    "obj-001",
				ObjectType:  enums.WorkflowObjectType("risk"),
			},
		}

		got := buildIngestMetadata(integration, "collect", mappedIngestRecord{Schema: "risk"}, options)

		if got.WorkflowInstanceID != "wf-001" {
			t.Fatalf("WorkflowInstanceID=%q, want %q", got.WorkflowInstanceID, "wf-001")
		}
		if got.WorkflowActionKey != "action-key" {
			t.Fatalf("WorkflowActionKey=%q, want %q", got.WorkflowActionKey, "action-key")
		}
		if got.WorkflowActionIndex != 3 {
			t.Fatalf("WorkflowActionIndex=%d, want %d", got.WorkflowActionIndex, 3)
		}
		if got.WorkflowObjectID != "obj-001" {
			t.Fatalf("WorkflowObjectID=%q, want %q", got.WorkflowObjectID, "obj-001")
		}
		if got.WorkflowObjectType != "risk" {
			t.Fatalf("WorkflowObjectType=%q, want %q", got.WorkflowObjectType, "risk")
		}
	})
}

func TestBuildIngestHeaders(t *testing.T) {
	t.Parallel()

	t.Run("includes non-empty properties and tags", func(t *testing.T) {
		t.Parallel()

		record := mappedIngestRecord{Schema: "finding", Variant: "cve"}
		metadata := integrationgenerated.IntegrationIngestMetadata{
			IntegrationID: "int-001",
			DefinitionID:  "def-001",
			Operation:     "collect",
			Source:        integrationgenerated.IntegrationIngestSourceWebhook,
			RunID:         "run-001",
			Variant:       "cve",
		}

		headers := buildIngestHeaders(record, metadata)

		if headers.Properties["schema"] != "finding" {
			t.Fatalf("schema property=%q, want %q", headers.Properties["schema"], "finding")
		}
		if headers.Properties["integration_id"] != "int-001" {
			t.Fatalf("integration_id property=%q, want %q", headers.Properties["integration_id"], "int-001")
		}
		if _, ok := headers.Properties["delivery_id"]; ok {
			t.Fatal("expected delivery_id to be omitted when empty")
		}
		if len(headers.Tags) != 2 {
			t.Fatalf("expected 2 tags, got %d", len(headers.Tags))
		}
	})

	t.Run("omits empty source from tags", func(t *testing.T) {
		t.Parallel()

		record := mappedIngestRecord{Schema: "asset"}
		metadata := integrationgenerated.IntegrationIngestMetadata{}

		headers := buildIngestHeaders(record, metadata)

		if len(headers.Tags) != 1 {
			t.Fatalf("expected 1 tag, got %d: %v", len(headers.Tags), headers.Tags)
		}
		if headers.Tags[0] != "asset" {
			t.Fatalf("expected tag %q, got %q", "asset", headers.Tags[0])
		}
	})
}

func TestPersistMappedRecord_UnsupportedSchema(t *testing.T) {
	t.Parallel()

	err := persistMappedRecord(context.Background(), nil, nil, "nonexistent_schema", json.RawMessage(`{}`))
	if !errors.Is(err, ErrIngestUnsupportedSchema) {
		t.Fatalf("expected ErrIngestUnsupportedSchema, got %v", err)
	}
}

func TestEmitMappedRecord_UnsupportedSchema(t *testing.T) {
	t.Parallel()

	record := mappedIngestRecord{Schema: "nonexistent_schema"}
	err := emitMappedRecord(context.Background(), nil, nil, "op", record, IngestOptions{})
	if !errors.Is(err, ErrIngestUnsupportedSchema) {
		t.Fatalf("expected ErrIngestUnsupportedSchema, got %v", err)
	}
}

func TestPrepareContactInput_SetsOwnerID(t *testing.T) {
	t.Parallel()

	integration := &ent.Integration{
		OwnerID: "org-001",
		Config:  openapi.IntegrationConfig{},
	}

	input := ent.CreateContactInput{}
	got := prepareContactInput(context.Background(), input, integration)

	if got.OwnerID == nil || *got.OwnerID != "org-001" {
		t.Fatalf("expected OwnerID=%q, got %v", "org-001", got.OwnerID)
	}
}

func TestPrepareContactInput_DoesNotOverrideOwnerID(t *testing.T) {
	t.Parallel()

	existing := "org-existing"
	integration := &ent.Integration{
		OwnerID: "org-001",
	}

	input := ent.CreateContactInput{OwnerID: &existing}
	got := prepareContactInput(context.Background(), input, integration)

	if *got.OwnerID != "org-existing" {
		t.Fatalf("expected OwnerID=%q, got %q", "org-existing", *got.OwnerID)
	}
}

func TestPrepareDirectoryAccountInput_SetsContextFields(t *testing.T) {
	t.Parallel()

	ctx := withDirectorySyncRunID(context.Background(), "dsrun-001")
	integration := &ent.Integration{
		OwnerID: "org-001",
	}

	input := ent.CreateDirectoryAccountInput{}
	got := prepareDirectoryAccountInput(ctx, input, integration)

	if got.OwnerID == nil || *got.OwnerID != "org-001" {
		t.Fatalf("expected OwnerID=%q, got %v", "org-001", got.OwnerID)
	}
	if got.DirectorySyncRunID == nil || *got.DirectorySyncRunID != "dsrun-001" {
		t.Fatalf("expected DirectorySyncRunID=%q, got %v", "dsrun-001", got.DirectorySyncRunID)
	}
}

func TestPrepareDirectoryGroupInput_SetsContextFields(t *testing.T) {
	t.Parallel()

	ctx := withDirectorySyncRunID(context.Background(), "dsrun-002")
	integration := &ent.Integration{
		OwnerID: "org-002",
	}

	input := ent.CreateDirectoryGroupInput{}
	got := prepareDirectoryGroupInput(ctx, input, integration)

	if got.OwnerID == nil || *got.OwnerID != "org-002" {
		t.Fatalf("expected OwnerID=%q, got %v", "org-002", got.OwnerID)
	}
	if got.DirectorySyncRunID != "dsrun-002" {
		t.Fatalf("expected DirectorySyncRunID=%q, got %q", "dsrun-002", got.DirectorySyncRunID)
	}
}

func TestPrepareDirectoryMembershipInput_SetsContextFields(t *testing.T) {
	t.Parallel()

	ctx := withDirectorySyncRunID(context.Background(), "dsrun-003")
	integration := &ent.Integration{
		OwnerID: "org-003",
	}

	input := ent.CreateDirectoryMembershipInput{}
	got := prepareDirectoryMembershipInput(ctx, input, integration)

	if got.OwnerID == nil || *got.OwnerID != "org-003" {
		t.Fatalf("expected OwnerID=%q, got %v", "org-003", got.OwnerID)
	}
	if got.DirectorySyncRunID != "dsrun-003" {
		t.Fatalf("expected DirectorySyncRunID=%q, got %q", "dsrun-003", got.DirectorySyncRunID)
	}
}

func TestRegisterIngestListeners_NilRuntime(t *testing.T) {
	t.Parallel()

	err := RegisterIngestListeners(nil)
	if !errors.Is(err, ErrGalaRequired) {
		t.Fatalf("expected ErrGalaRequired, got %v", err)
	}
}

func TestPrepareAssetInput(t *testing.T) {
	t.Parallel()

	integration := &ent.Integration{
		OwnerID: "org-001",
	}

	input := ent.CreateAssetInput{}
	got := prepareAssetInput(context.Background(), input, integration)

	// prepareAssetInput delegates to integrationgenerated.PrepareAssetInput
	// verify it returns without panic
	_ = got
}

func TestPrepareEntityInput(t *testing.T) {
	t.Parallel()

	integration := &ent.Integration{
		OwnerID: "org-001",
	}

	input := ent.CreateEntityInput{}
	got := prepareEntityInput(context.Background(), input, integration)
	_ = got
}

func TestPrepareFindingInput(t *testing.T) {
	t.Parallel()

	integration := &ent.Integration{
		OwnerID: "org-001",
	}

	input := ent.CreateFindingInput{}
	got := prepareFindingInput(context.Background(), input, integration)
	_ = got
}

func TestPrepareRiskInput(t *testing.T) {
	t.Parallel()

	integration := &ent.Integration{
		OwnerID: "org-001",
	}

	input := ent.CreateRiskInput{}
	got := prepareRiskInput(context.Background(), input, integration)
	_ = got
}

func TestPrepareVulnerabilityInput(t *testing.T) {
	t.Parallel()

	integration := &ent.Integration{
		OwnerID: "org-001",
	}

	input := ent.CreateVulnerabilityInput{}
	got := prepareVulnerabilityInput(context.Background(), input, integration)
	_ = got
}

func TestPrepareDirectoryAccountInput_NoOverrideWhenSet(t *testing.T) {
	t.Parallel()

	existing := "org-existing"
	existingRunID := "existing-run"
	integration := &ent.Integration{
		OwnerID: "org-001",
	}

	ctx := withDirectorySyncRunID(context.Background(), "dsrun-new")
	input := ent.CreateDirectoryAccountInput{
		OwnerID:            &existing,
		DirectorySyncRunID: &existingRunID,
	}
	got := prepareDirectoryAccountInput(ctx, input, integration)

	if *got.OwnerID != "org-existing" {
		t.Fatalf("expected OwnerID=%q, got %q", "org-existing", *got.OwnerID)
	}
	if *got.DirectorySyncRunID != "existing-run" {
		t.Fatalf("expected DirectorySyncRunID=%q, got %q", "existing-run", *got.DirectorySyncRunID)
	}
}

func TestPrepareDirectoryGroupInput_NoOverrideWhenSet(t *testing.T) {
	t.Parallel()

	existing := "org-existing"
	integration := &ent.Integration{
		OwnerID: "org-001",
	}

	ctx := withDirectorySyncRunID(context.Background(), "dsrun-new")
	input := ent.CreateDirectoryGroupInput{
		OwnerID:            &existing,
		DirectorySyncRunID: "existing-run",
	}
	got := prepareDirectoryGroupInput(ctx, input, integration)

	if *got.OwnerID != "org-existing" {
		t.Fatalf("expected OwnerID=%q, got %q", "org-existing", *got.OwnerID)
	}
	if got.DirectorySyncRunID != "existing-run" {
		t.Fatalf("expected DirectorySyncRunID=%q, got %q", "existing-run", got.DirectorySyncRunID)
	}
}

func TestPrepareDirectoryMembershipInput_NoOverrideWhenSet(t *testing.T) {
	t.Parallel()

	existing := "org-existing"
	integration := &ent.Integration{
		OwnerID: "org-001",
	}

	ctx := withDirectorySyncRunID(context.Background(), "dsrun-new")
	input := ent.CreateDirectoryMembershipInput{
		OwnerID:            &existing,
		DirectorySyncRunID: "existing-run",
	}
	got := prepareDirectoryMembershipInput(ctx, input, integration)

	if *got.OwnerID != "org-existing" {
		t.Fatalf("expected OwnerID=%q, got %q", "org-existing", *got.OwnerID)
	}
	if got.DirectorySyncRunID != "existing-run" {
		t.Fatalf("expected DirectorySyncRunID=%q, got %q", "existing-run", got.DirectorySyncRunID)
	}
}
