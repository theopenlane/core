package types //nolint:revive

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// runtimeTestConfig is the runtime config type the runtime ref tests reflect
type runtimeTestConfig struct {
	Endpoint string `json:"endpoint"`
}

func TestNewRuntimeRefReflectsSchema(t *testing.T) {
	t.Parallel()

	ref := NewRuntimeRef[runtimeTestConfig]("runtime")

	if ref.ID() != NewRuntimeRefID("runtime") {
		t.Fatalf("ID() = %q", ref.ID())
	}

	if jsonx.SchemaID(ref.Schema()) != "runtimeTestConfig" {
		t.Fatalf("expected the reflected schema retained, got %s", ref.Schema())
	}
}

func TestRuntimeRefOfDerivesNameFromSchema(t *testing.T) {
	t.Parallel()

	ref := RuntimeRefOf[runtimeTestConfig]()

	if !ref.ID().Valid() || ref.ID().Name() != "runtimeTestConfig" {
		t.Fatalf("ID() = %q", ref.ID())
	}
}

func TestRuntimeRefRegistration(t *testing.T) {
	t.Parallel()

	ref := RuntimeRefOf[runtimeTestConfig]()

	reg := ref.Registration(RuntimeIntegrationRegistration{
		Config: json.RawMessage(`{"endpoint":"x"}`),
		Build:  func(context.Context, json.RawMessage) (any, error) { return "built", nil },
	})

	if reg.Ref != ref.ID() {
		t.Fatalf("Ref = %q, want %q", reg.Ref, ref.ID())
	}

	if string(reg.Schema) != string(ref.Schema()) {
		t.Fatalf("Schema = %s", reg.Schema)
	}

	if string(reg.Config) != `{"endpoint":"x"}` || reg.Build == nil {
		t.Fatalf("expected base fields preserved, got %+v", reg)
	}
}
