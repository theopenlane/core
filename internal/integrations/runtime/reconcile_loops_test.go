package runtime

import (
	"encoding/json"
	"testing"

	"github.com/theopenlane/core/common/enums"
)

// TestReconcileLoopFragment verifies the fragment carries installation, operation, and run type
func TestReconcileLoopFragment(t *testing.T) {
	t.Parallel()

	fragment, err := reconcileLoopFragment("install-1", "sync.users")
	if err != nil {
		t.Fatalf("reconcileLoopFragment() error = %v", err)
	}

	var wire struct {
		Properties map[string]string `json:"properties"`
	}

	if err := json.Unmarshal([]byte(fragment), &wire); err != nil {
		t.Fatalf("unmarshal fragment: %v", err)
	}

	if wire.Properties["entityId"] != "install-1" {
		t.Fatalf("entityId = %q, want %q", wire.Properties["entityId"], "install-1")
	}

	if wire.Properties["operation"] != "sync.users" {
		t.Fatalf("operation = %q, want %q", wire.Properties["operation"], "sync.users")
	}

	if wire.Properties["runType"] != enums.IntegrationRunTypeReconcile.String() {
		t.Fatalf("runType = %q, want %q", wire.Properties["runType"], enums.IntegrationRunTypeReconcile.String())
	}
}
