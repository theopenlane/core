package operations

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/theopenlane/core/common/enums"
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

func TestLookupIngestSchema(t *testing.T) {
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

			_, ok := lookupIngestSchema(tc.schema)
			if ok != tc.want {
				t.Fatalf("lookupIngestSchema(%q)=%v, want %v", tc.schema, ok, tc.want)
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

func TestPersistMappedRecord_UnsupportedSchema(t *testing.T) {
	t.Parallel()

	_, err := persistMappedRecord(context.Background(), nil, nil, "nonexistent_schema", json.RawMessage(`{}`))
	if !errors.Is(err, ErrIngestUnsupportedSchema) {
		t.Fatalf("expected ErrIngestUnsupportedSchema, got %v", err)
	}
}

func TestPrepareDirectoryAccountInput_SetsRunID(t *testing.T) {
	t.Parallel()

	ctx := withDirectorySyncRunID(context.Background(), "dsrun-001")

	got := prepareDirectoryAccountInput(ctx, ent.CreateDirectoryAccountInput{})

	if got.DirectorySyncRunID == nil || *got.DirectorySyncRunID != "dsrun-001" {
		t.Fatalf("expected DirectorySyncRunID=%q, got %v", "dsrun-001", got.DirectorySyncRunID)
	}
}

func TestPrepareDirectoryGroupInput_SetsRunID(t *testing.T) {
	t.Parallel()

	ctx := withDirectorySyncRunID(context.Background(), "dsrun-002")

	got := prepareDirectoryGroupInput(ctx, ent.CreateDirectoryGroupInput{})

	if got.DirectorySyncRunID != "dsrun-002" {
		t.Fatalf("expected DirectorySyncRunID=%q, got %q", "dsrun-002", got.DirectorySyncRunID)
	}
}

func TestPrepareDirectoryMembershipInput_SetsRunID(t *testing.T) {
	t.Parallel()

	ctx := withDirectorySyncRunID(context.Background(), "dsrun-003")

	got := prepareDirectoryMembershipInput(ctx, ent.CreateDirectoryMembershipInput{})

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

func TestPrepareDirectoryAccountInput_NoOverrideWhenSet(t *testing.T) {
	t.Parallel()

	existingRunID := "existing-run"
	ctx := withDirectorySyncRunID(context.Background(), "dsrun-new")

	got := prepareDirectoryAccountInput(ctx, ent.CreateDirectoryAccountInput{
		DirectorySyncRunID: &existingRunID,
	})

	if *got.DirectorySyncRunID != "existing-run" {
		t.Fatalf("expected DirectorySyncRunID=%q, got %q", "existing-run", *got.DirectorySyncRunID)
	}
}

func TestPrepareDirectoryGroupInput_NoOverrideWhenSet(t *testing.T) {
	t.Parallel()

	ctx := withDirectorySyncRunID(context.Background(), "dsrun-new")

	got := prepareDirectoryGroupInput(ctx, ent.CreateDirectoryGroupInput{
		DirectorySyncRunID: "existing-run",
	})

	if got.DirectorySyncRunID != "existing-run" {
		t.Fatalf("expected DirectorySyncRunID=%q, got %q", "existing-run", got.DirectorySyncRunID)
	}
}

func TestPrepareDirectoryMembershipInput_NoOverrideWhenSet(t *testing.T) {
	t.Parallel()

	ctx := withDirectorySyncRunID(context.Background(), "dsrun-new")

	got := prepareDirectoryMembershipInput(ctx, ent.CreateDirectoryMembershipInput{
		DirectorySyncRunID: "existing-run",
	})

	if got.DirectorySyncRunID != "existing-run" {
		t.Fatalf("expected DirectorySyncRunID=%q, got %q", "existing-run", got.DirectorySyncRunID)
	}
}
