package registry

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"

	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// finalizeConfig is a test stored-input config type with one specific key beside the uniform settings
type finalizeConfig struct {
	integrationtypes.OperationSettings
	// Limit bounds the number of records the operation reads
	Limit int `json:"limit,omitempty"`
}

// finalizeEmptyConfig carries only the uniform settings
type finalizeEmptyConfig struct {
	integrationtypes.OperationSettings
}

// finalizePayload is a caller-supplied payload type with no stored input
type finalizePayload struct {
	// Target names what the payload operation acts on
	Target string `json:"target"`
}

// finalizeDefinitionRef is the definition identity for finalize tests
var finalizeDefinitionRef = integrationtypes.NewDefinitionRef("def_finalize")

// finalizeOperation returns a reconciled stored-input operation registration over Cfg
func finalizeOperation[Cfg integrationtypes.OperationInput]() integrationtypes.OperationRegistration {
	return integrationtypes.OperationRefOf[Cfg]().
		Policy(integrationtypes.ExecutionPolicy{Reconcile: true}).
		HandlesRequest(func(context.Context, integrationtypes.OperationRequest, Cfg) (json.RawMessage, error) {
			return nil, nil
		}).
		Registration()
}

// operationDefinition returns a definition with the given operations
func operationDefinition(operations ...integrationtypes.OperationRegistration) integrationtypes.Definition {
	return integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: finalizeDefinitionRef.ID()},
		Operations:     operations,
	}
}

// propertyKeys lists a schema root's property keys in declaration order
func propertyKeys(t *testing.T, schema json.RawMessage) []string {
	t.Helper()

	root, _, err := jsonx.SchemaRoot(schema)
	if err != nil {
		t.Fatalf("SchemaRoot() error = %v", err)
	}

	var keys []string
	for pair := root.Properties.Oldest(); pair != nil; pair = pair.Next() {
		keys = append(keys, pair.Key)
	}

	return keys
}

// TestFinalizeKeepsEmbeddedOperationInputSchema verifies a stored-input config's embedded settings lead its one schema
func TestFinalizeKeepsEmbeddedOperationInputSchema(t *testing.T) {
	t.Parallel()

	def := operationDefinition(finalizeOperation[finalizeConfig](), finalizeOperation[finalizeEmptyConfig]())

	finalized := finalizeDefinition(def)

	configured := finalized.Operations[0]

	if got := propertyKeys(t, configured.Input.Schema); !slices.Equal(got, []string{"disable", "filterExpr", "limit"}) {
		t.Fatalf("input properties = %v, want the embedded settings followed by the config keys", got)
	}

	if !configured.Stored {
		t.Fatal("expected the operation input stored per installation")
	}

	if configured.Input.Name != "finalizeConfig" {
		t.Fatalf("Input.Name = %q", configured.Input.Name)
	}

	empty := finalized.Operations[1]

	if got := propertyKeys(t, empty.Input.Schema); !slices.Equal(got, []string{"disable", "filterExpr"}) {
		t.Fatalf("empty config input properties = %v, want only the settings", got)
	}

	if def.Operations[0].Input.Validate != nil {
		t.Fatal("expected finalize to leave the builder's operations untouched")
	}

	stored := json.RawMessage(`{"disable":true,"filterExpr":"payload.ok","limit":3}`)

	result, err := jsonx.ValidateSchema(configured.Input.Schema, stored)
	if err != nil || !result.Valid() {
		t.Fatalf("expected the stored input schema to accept a stored document, got %v %v", err, jsonx.ValidationErrorStrings(result))
	}

	if !configured.DisabledFor(stored) {
		t.Fatal("expected the stored disable key to switch the operation off")
	}

	if configured.DisabledFor(json.RawMessage(`{"limit":3}`)) {
		t.Fatal("expected a stored document without the disable key to leave the operation on")
	}
}

// TestFinalizeValidatesOperationFilterExpr verifies the composed input rejects a filter expression that does not compile and still runs the definition's own check
func TestFinalizeValidatesOperationFilterExpr(t *testing.T) {
	t.Parallel()

	errLimit := errors.New("limit too high")

	limited := integrationtypes.OperationRefOf[finalizeConfig]().
		Policy(integrationtypes.ExecutionPolicy{Reconcile: true}).
		HandlesRequest(func(context.Context, integrationtypes.OperationRequest, finalizeConfig) (json.RawMessage, error) {
			return nil, nil
		}).
		Validated(func(_ context.Context, _ integrationtypes.InstallationRequest, config *finalizeConfig) error {
			if config.Limit > 10 {
				return errLimit
			}

			return nil
		}).
		Registration()

	finalized := finalizeDefinition(operationDefinition(limited, finalizeOperation[finalizeEmptyConfig]()))

	tests := []struct {
		name      string
		operation integrationtypes.OperationRegistration
		payload   string
		wantErr   error
	}{
		{name: "valid expression passes", operation: finalized.Operations[0], payload: `{"filterExpr":"payload.ok == true","limit":3}`},
		{name: "empty expression passes", operation: finalized.Operations[0], payload: `{"limit":3}`},
		{name: "expression that does not compile is rejected", operation: finalized.Operations[0], payload: `{"filterExpr":"payload.","limit":3}`, wantErr: ErrOperationFilterExprInvalid},
		{name: "definition validation runs after the framework check", operation: finalized.Operations[0], payload: `{"filterExpr":"true","limit":11}`, wantErr: errLimit},
		{name: "operation without its own validation still compiles the expression", operation: finalized.Operations[1], payload: `{"filterExpr":"payload."}`, wantErr: ErrOperationFilterExprInvalid},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.operation.Input.Validate(context.Background(), integrationtypes.InstallationRequest{}, json.RawMessage(tc.payload))

			switch {
			case tc.wantErr == nil && err != nil:
				t.Fatalf("Validate() error = %v, want nil", err)
			case tc.wantErr != nil && !errors.Is(err, tc.wantErr):
				t.Fatalf("Validate() error = %v, want %v", err, tc.wantErr)
			}
		})
	}
}

// TestFinalizeLeavesOperationsWithoutInput verifies literal and payload operations stay without a stored input
func TestFinalizeLeavesOperationsWithoutInput(t *testing.T) {
	t.Parallel()

	payload := integrationtypes.OperationPayloadOf[finalizePayload]().
		HandlesRequest(func(context.Context, integrationtypes.OperationRequest, finalizePayload) (json.RawMessage, error) {
			return nil, nil
		}).
		Registration()

	def := operationDefinition(integrationtypes.OperationRegistration{Name: "literal", Handle: newTestHandler()}, payload)

	finalized := finalizeDefinition(def)

	for _, operation := range finalized.Operations {
		if operation.Stored || operation.Input.Validate != nil {
			t.Fatalf("expected no stored input for operation %s, got %+v", operation.Name, operation.Input)
		}
	}

	if got := propertyKeys(t, finalized.Operations[1].Input.Schema); !slices.Equal(got, []string{"target"}) {
		t.Fatalf("payload config properties = %v, want only the payload keys", got)
	}
}
