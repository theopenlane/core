package state

import "testing"

func TestProviderDataReturnsDeepClone(t *testing.T) {
	s := IntegrationProviderState{}
	changed, err := s.MergeProviderData("githubapp", map[string]any{
		"installationId": "123",
		"nested": map[string]any{
			"enabled": true,
		},
	})
	if err != nil {
		t.Fatalf("expected merge success: %v", err)
	}
	if !changed {
		t.Fatalf("expected merge to change state")
	}

	out, err := s.ProviderDataMap("githubapp")
	if err != nil {
		t.Fatalf("expected provider data: %v", err)
	}
	if out == nil {
		t.Fatalf("expected provider data")
	}

	nested, ok := out["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map")
	}
	nested["enabled"] = false

	source, err := s.ProviderDataMap("githubapp")
	if err != nil {
		t.Fatalf("expected provider data map: %v", err)
	}
	sourceNested, ok := source["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected source nested map")
	}
	if sourceNested["enabled"] != true {
		t.Fatalf("expected source nested value to remain true")
	}
}

func TestMergeProviderDataMergesDeeply(t *testing.T) {
	s := IntegrationProviderState{}
	_, err := s.MergeProviderData("githubapp", map[string]any{
		"appId": "10",
		"nested": map[string]any{
			"a": "one",
		},
	})
	if err != nil {
		t.Fatalf("expected initial merge success: %v", err)
	}

	changed, err := s.MergeProviderData("githubapp", map[string]any{
		"installationId": "20",
		"nested": map[string]any{
			"b": "two",
		},
	})
	if err != nil {
		t.Fatalf("expected merge success: %v", err)
	}
	if !changed {
		t.Fatalf("expected merge to change state")
	}

	provider, err := s.ProviderDataMap("githubapp")
	if err != nil {
		t.Fatalf("expected provider data map: %v", err)
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
	_, err := s.MergeProviderData("slack", map[string]any{
		"teamId": "T123",
	})
	if err != nil {
		t.Fatalf("expected initial merge success: %v", err)
	}

	changed, err := s.MergeProviderData("slack", map[string]any{
		"teamId": "T123",
	})
	if err != nil {
		t.Fatalf("expected merge success: %v", err)
	}
	if changed {
		t.Fatalf("expected no change")
	}
}
