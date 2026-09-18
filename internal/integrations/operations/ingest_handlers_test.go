package operations

import (
	"testing"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/types"
)

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
