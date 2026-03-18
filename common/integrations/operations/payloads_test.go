package operations

import "testing"

func TestAddPayloadIf(t *testing.T) {
	original := map[string]any{"a": 1}
	if out := AddPayloadIf(original, false, "payload", "value"); out["payload"] != nil {
		if _, ok := out["payload"]; ok {
			t.Fatalf("expected payload not to be added")
		}
	}

	out := AddPayloadIf(nil, true, "payload", "value")
	if out["payload"] != "value" {
		t.Fatalf("expected payload to be added")
	}
}
