package engine

import (
	"encoding/json"
	"testing"
)

func TestDecodeApprovalActionParamsSupportsLegacyAssignees(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{
		"assignees": {
			"users": ["user-1"],
			"groups": ["group-1"]
		},
		"fields": ["status"]
	}`)

	params, err := decodeApprovalActionParams(raw)
	if err != nil {
		t.Fatalf("decodeApprovalActionParams() unexpected error: %v", err)
	}

	if len(params.Targets) != 2 {
		t.Fatalf("expected 2 targets from legacy assignees, got %d", len(params.Targets))
	}

	if params.Targets[0].ID != "user-1" || params.Targets[1].ID != "group-1" {
		t.Fatalf("unexpected target conversion: %#v", params.Targets)
	}
}

func TestDecodeApprovalActionParamsCoercesLegacyRequired(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{
		"targets": [{"type":"USER","id":"user-1"}],
		"required": "false"
	}`)

	params, err := decodeApprovalActionParams(raw)
	if err != nil {
		t.Fatalf("decodeApprovalActionParams() unexpected error: %v", err)
	}
	if params.Required == nil || *params.Required {
		t.Fatalf("expected required=false, got %#v", params.Required)
	}

	raw = json.RawMessage(`{
		"targets": [{"type":"USER","id":"user-1"}],
		"required": 2
	}`)

	params, err = decodeApprovalActionParams(raw)
	if err != nil {
		t.Fatalf("decodeApprovalActionParams() number unexpected error: %v", err)
	}
	if params.Required == nil || !*params.Required {
		t.Fatalf("expected required=true, got %#v", params.Required)
	}
	if params.RequiredCount != 2 {
		t.Fatalf("expected required_count to inherit legacy value 2, got %d", params.RequiredCount)
	}
}

func TestDecodeApprovalActionParamsRejectsInvalidRequired(t *testing.T) {
	t.Parallel()

	raw := json.RawMessage(`{
		"targets": [{"type":"USER","id":"user-1"}],
		"required": -1
	}`)

	if _, err := decodeApprovalActionParams(raw); err == nil {
		t.Fatalf("expected invalid required to fail decode")
	}
}
