package operations

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/entityops"
	ent "github.com/theopenlane/core/internal/ent/generated"
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

func TestLookupIngestHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		schema string
		want   bool
	}{
		{"asset schema exists", entityops.SchemaAsset.Name, true},
		{"contact schema exists", entityops.SchemaContact.Name, true},
		{"directory account exists", entityops.SchemaDirectoryAccount.Name, true},
		{"directory group exists", entityops.SchemaDirectoryGroup.Name, true},
		{"directory membership exists", entityops.SchemaDirectoryMembership.Name, true},
		{"entity exists", entityops.SchemaEntity.Name, true},
		{"finding exists", entityops.SchemaFinding.Name, true},
		{"risk exists", entityops.SchemaRisk.Name, true},
		{"vulnerability exists", entityops.SchemaVulnerability.Name, true},
		{"unknown schema missing", "nonexistent", false},
		{"empty string missing", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, ok := lookupIngestHandler(tc.schema)
			if ok != tc.want {
				t.Fatalf("lookupIngestHandler(%q)=%v, want %v", tc.schema, ok, tc.want)
			}
		})
	}
}

func TestBuildIngestOperationContext(t *testing.T) {
	t.Parallel()

	integration := &ent.Integration{
		ID:           "int-001",
		DefinitionID: "def-001",
		OwnerID:      "org-001",
	}

	t.Run("promotes integration and carries provenance", func(t *testing.T) {
		t.Parallel()

		options := IngestOptions{
			RunID:        "run-001",
			Webhook:      "github",
			WebhookEvent: "push",
			DeliveryID:   "delivery-001",
		}

		oc := buildIngestOperationContext(integration, options)

		if oc.OwnerID != "org-001" {
			t.Fatalf("OwnerID=%q, want %q", oc.OwnerID, "org-001")
		}
		if oc.EntityID != "int-001" {
			t.Fatalf("EntityID=%q, want %q", oc.EntityID, "int-001")
		}

		src := types.IntegrationSourceFrom(oc)
		if src.IntegrationID != "int-001" {
			t.Fatalf("IntegrationID=%q, want %q", src.IntegrationID, "int-001")
		}
		if src.DefinitionID != "def-001" {
			t.Fatalf("DefinitionID=%q, want %q", src.DefinitionID, "def-001")
		}
		if src.RunID != "run-001" {
			t.Fatalf("RunID=%q, want %q", src.RunID, "run-001")
		}
		if src.Webhook != "github" {
			t.Fatalf("Webhook=%q, want %q", src.Webhook, "github")
		}
	})

	t.Run("includes workflow provenance", func(t *testing.T) {
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

		oc := buildIngestOperationContext(integration, options)
		src := types.IntegrationSourceFrom(oc)

		if src.Workflow == nil || src.Workflow.InstanceID != "wf-001" {
			t.Fatalf("Workflow.InstanceID=%v, want %q", src.Workflow, "wf-001")
		}
	})
}

func TestBuildIngestHeaders(t *testing.T) {
	t.Parallel()

	integration := &ent.Integration{ID: "int-001", DefinitionID: "def-001"}

	t.Run("includes non-empty properties and tags", func(t *testing.T) {
		t.Parallel()

		record := mappedIngestRecord{Schema: "finding", Variant: "cve"}
		options := IngestOptions{
			RunID: "run-001",
		}

		headers := buildIngestHeaders(integration, "collect", record, options)

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

	t.Run("tags definition and schema", func(t *testing.T) {
		t.Parallel()

		record := mappedIngestRecord{Schema: "asset"}

		headers := buildIngestHeaders(integration, "", record, IngestOptions{})

		if !slices.Contains(headers.Tags, "schema_asset") {
			t.Fatalf("expected tag %q, got %v", "schema_asset", headers.Tags)
		}
		if !slices.Contains(headers.Tags, "def-001") {
			t.Fatalf("expected tag %q, got %v", "def-001", headers.Tags)
		}
	})
}

func TestPersistMappedRecord_UnsupportedSchema(t *testing.T) {
	t.Parallel()

	_, err := persistMappedRecord(context.Background(), nil, nil, "nonexistent_schema", json.RawMessage(`{}`))
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
