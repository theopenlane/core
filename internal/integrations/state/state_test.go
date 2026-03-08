package state

import (
	"encoding/json"
	"testing"
)

func TestProviderDataReturnsRawMessage(t *testing.T) {
	s := IntegrationProviderState{}
	patch, err := json.Marshal(map[string]any{
		"installationId": "123",
		"nested": map[string]any{
			"enabled": true,
		},
	})
	if err != nil {
		t.Fatalf("expected marshal success: %v", err)
	}

	changed, err := s.MergeProviderData("githubapp", patch)
	if err != nil {
		t.Fatalf("expected merge success: %v", err)
	}
	if !changed {
		t.Fatalf("expected merge to change state")
	}

	raw := s.ProviderData("githubapp")
	if len(raw) == 0 {
		t.Fatalf("expected provider data")
	}

	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("expected unmarshal success: %v", err)
	}
	if out["installationId"] != "123" {
		t.Fatalf("expected installationId to be '123'")
	}
}

func TestMergeProviderDataMergesDeeply(t *testing.T) {
	s := IntegrationProviderState{}

	patch1, err := json.Marshal(map[string]any{
		"appId": "10",
		"nested": map[string]any{
			"a": "one",
		},
	})
	if err != nil {
		t.Fatalf("expected marshal success: %v", err)
	}

	if _, err := s.MergeProviderData("githubapp", patch1); err != nil {
		t.Fatalf("expected initial merge success: %v", err)
	}

	patch2, err := json.Marshal(map[string]any{
		"installationId": "20",
		"nested": map[string]any{
			"b": "two",
		},
	})
	if err != nil {
		t.Fatalf("expected marshal success: %v", err)
	}

	changed, err := s.MergeProviderData("githubapp", patch2)
	if err != nil {
		t.Fatalf("expected merge success: %v", err)
	}
	if !changed {
		t.Fatalf("expected merge to change state")
	}

	raw := s.ProviderData("githubapp")
	var provider map[string]any
	if err := json.Unmarshal(raw, &provider); err != nil {
		t.Fatalf("expected unmarshal success: %v", err)
	}
	if provider["appId"] != "10" {
		t.Fatalf("expected appId to be preserved")
	}
	if provider["installationId"] != "20" {
		t.Fatalf("expected installationId to be set")
	}
	nested, ok := provider["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map")
	}
	if nested["a"] != "one" || nested["b"] != "two" {
		t.Fatalf("expected nested deep merge result")
	}
}

func TestMergeProviderDataNoChange(t *testing.T) {
	s := IntegrationProviderState{}

	patch, err := json.Marshal(map[string]any{
		"teamId": "T123",
	})
	if err != nil {
		t.Fatalf("expected marshal success: %v", err)
	}

	if _, err := s.MergeProviderData("slack", patch); err != nil {
		t.Fatalf("expected initial merge success: %v", err)
	}

	changed, err := s.MergeProviderData("slack", patch)
	if err != nil {
		t.Fatalf("expected merge success: %v", err)
	}
	if changed {
		t.Fatalf("expected no change")
	}
}
