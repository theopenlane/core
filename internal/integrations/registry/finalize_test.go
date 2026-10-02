package registry

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/samber/lo"

	integrationtypes "github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
)

// finalizeConfig is a test operation config type with one specific key
type finalizeConfig struct {
	// Limit bounds the number of records the operation reads
	Limit int `json:"limit,omitempty"`
}

// finalizeReservedConfig declares a key the uniform operation settings reserve
type finalizeReservedConfig struct {
	// Disable collides with the uniform disable key
	Disable bool `json:"disable,omitempty"`
}

// finalizeEmptyConfig has no properties
type finalizeEmptyConfig struct{}

// finalizeDefinitionRef is the definition identity for finalize tests
var finalizeDefinitionRef = integrationtypes.NewDefinitionRef("def_finalize")

// finalizeOperation returns a reconciled operation registration over Cfg
func finalizeOperation[Cfg any]() integrationtypes.OperationRegistration {
	return integrationtypes.OperationRefOf[Cfg]().
		Policy(integrationtypes.ExecutionPolicy{Reconcile: true}).
		HandlesRequest(func(context.Context, integrationtypes.OperationRequest, Cfg) (json.RawMessage, error) {
			return nil, nil
		}).
		Registration(finalizeDefinitionRef, integrationtypes.OperationRegistration{})
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

// sameJSON reports whether two raw documents decode to the same value regardless of encoding
func sameJSON(a, b json.RawMessage) bool {
	var left, right any

	if err := json.Unmarshal(a, &left); err != nil {
		return false
	}

	if err := json.Unmarshal(b, &right); err != nil {
		return false
	}

	return reflect.DeepEqual(left, right)
}

// TestFinalizeCredentialFormSchema verifies form/stored schema derivation across slot kinds
func TestFinalizeCredentialFormSchema(t *testing.T) {
	t.Parallel()

	literalSchema := json.RawMessage(`{"type":"object"}`)
	literalSlot := integrationtypes.NewCredentialSlotID("literal")

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_finalize_credentials"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRef.Registration(integrationtypes.CredentialRegistration{}),
			testAuthCredentialRef.Registration(integrationtypes.CredentialRegistration{}),
			{Ref: literalSlot, Schema: literalSchema},
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{
				CredentialRef: testAuthCredentialRef.ID(),
				Auth:          &integrationtypes.AuthRegistration{CredentialRef: testAuthCredentialRef.ID()},
			},
		},
	}

	finalized, err := finalizeDefinition(def)
	if err != nil {
		t.Fatalf("finalizeDefinition() error = %v", err)
	}

	form, _ := finalized.CredentialRegistration(testCredentialRef.ID())
	if !sameJSON(form.Schema, testCredentialRef.Schema()) {
		t.Fatalf("form slot Schema = %s, want its stored schema", form.Schema)
	}

	authManaged, _ := finalized.CredentialRegistration(testAuthCredentialRef.ID())
	if len(authManaged.Schema) != 0 {
		t.Fatalf("auth-managed slot Schema = %s, want empty", authManaged.Schema)
	}

	if !sameJSON(authManaged.StoredSchema, testAuthCredentialRef.Schema()) {
		t.Fatalf("auth-managed slot StoredSchema = %s, want the reflected schema", authManaged.StoredSchema)
	}

	literal, _ := finalized.CredentialRegistration(literalSlot)
	if !sameJSON(literal.StoredSchema, literalSchema) || !sameJSON(literal.Schema, literalSchema) {
		t.Fatalf("literal slot = %+v, want stored and form schema from the literal", literal)
	}

	if len(def.CredentialRegistrations[0].Schema) != 0 {
		t.Fatal("expected finalize to leave the builder's registrations untouched")
	}
}

// TestFinalizeComposesOperationInputSchema verifies the uniform settings lead every operation's stored schema
func TestFinalizeComposesOperationInputSchema(t *testing.T) {
	t.Parallel()

	def := operationDefinition(finalizeOperation[finalizeConfig](), finalizeOperation[finalizeEmptyConfig]())

	finalized, err := finalizeDefinition(def)
	if err != nil {
		t.Fatalf("finalizeDefinition() error = %v", err)
	}

	configured := finalized.Operations[0]

	if got := propertyKeys(t, configured.Input.Schema); !slices.Equal(got, []string{"disable", "filterExpr", "limit"}) {
		t.Fatalf("input properties = %v, want the settings followed by the config keys", got)
	}

	if !sameJSON(configured.ConfigSchema, integrationtypes.OperationRefOf[finalizeConfig]().Schema()) {
		t.Fatalf("expected ConfigSchema to stay the pure config schema, got %s", configured.ConfigSchema)
	}

	if configured.Input.Name != "finalizeConfig" {
		t.Fatalf("Input.Name = %q", configured.Input.Name)
	}

	empty := finalized.Operations[1]

	if got := propertyKeys(t, empty.Input.Schema); !slices.Equal(got, []string{"disable", "filterExpr"}) {
		t.Fatalf("empty config input properties = %v, want only the settings", got)
	}

	if sameJSON(def.Operations[0].Input.Schema, configured.Input.Schema) {
		t.Fatal("expected finalize to leave the builder's operations untouched")
	}

	stored := json.RawMessage(`{"disable":true,"filterExpr":"payload.ok","limit":3}`)

	result, err := jsonx.ValidateSchema(configured.Input.Schema, stored)
	if err != nil || !result.Valid() {
		t.Fatalf("expected the composed schema to accept a stored document, got %v %v", err, jsonx.ValidationErrorStrings(result))
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
		Registration(finalizeDefinitionRef, integrationtypes.OperationRegistration{})

	finalized, err := finalizeDefinition(operationDefinition(limited, finalizeOperation[finalizeEmptyConfig]()))
	if err != nil {
		t.Fatalf("finalizeDefinition() error = %v", err)
	}

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

// TestFinalizeRejectsReservedConfigKeys verifies a config declaring a settings key fails registration
func TestFinalizeRejectsReservedConfigKeys(t *testing.T) {
	t.Parallel()

	if _, err := finalizeDefinition(operationDefinition(finalizeOperation[finalizeReservedConfig]())); !errors.Is(err, ErrOperationConfigReservedKey) {
		t.Fatalf("finalizeDefinition() error = %v, want %v", err, ErrOperationConfigReservedKey)
	}
}

// TestFinalizeLeavesLiteralOperationsWithoutInput verifies operations declared without a stored input stay bare
func TestFinalizeLeavesLiteralOperationsWithoutInput(t *testing.T) {
	t.Parallel()

	def := operationDefinition(integrationtypes.OperationRegistration{Name: "literal", Handle: newTestHandler()})

	finalized, err := finalizeDefinition(def)
	if err != nil {
		t.Fatalf("finalizeDefinition() error = %v", err)
	}

	if finalized.Operations[0].Input != nil {
		t.Fatalf("expected no stored input derived for a literal operation, got %+v", finalized.Operations[0].Input)
	}
}

// TestFinalizeDefaultsCredentialSlots verifies clients and connections without slots take the declared ones
func TestFinalizeDefaultsCredentialSlots(t *testing.T) {
	t.Parallel()

	clientRef := integrationtypes.ClientRefOf[string]()
	explicit := integrationtypes.NewClientRef[int]("explicit").Using(testAuthCredentialRef)
	build := func(context.Context, integrationtypes.ClientBuildRequest) (string, error) { return "", nil }
	buildInt := func(context.Context, integrationtypes.ClientBuildRequest) (int, error) { return 0, nil }

	def := integrationtypes.Definition{
		DefinitionSpec: integrationtypes.DefinitionSpec{ID: "def_finalize_slots"},
		CredentialRegistrations: []integrationtypes.CredentialRegistration{
			testCredentialRef.Registration(integrationtypes.CredentialRegistration{}),
			testAuthCredentialRef.Registration(integrationtypes.CredentialRegistration{}),
		},
		Clients: []integrationtypes.ClientRegistration{
			clientRef.Registration(build, integrationtypes.ClientRegistration{}),
			explicit.Registration(buildInt, integrationtypes.ClientRegistration{}),
		},
		Connections: []integrationtypes.ConnectionRegistration{
			{CredentialRef: testCredentialRef.ID()},
			{CredentialRef: testAuthCredentialRef.ID(), CredentialRefs: []integrationtypes.CredentialSlotID{testAuthCredentialRef.ID(), testCredentialRef.ID()}},
		},
	}

	finalized, err := finalizeDefinition(def)
	if err != nil {
		t.Fatalf("finalizeDefinition() error = %v", err)
	}

	declared := []integrationtypes.CredentialSlotID{testCredentialRef.ID(), testAuthCredentialRef.ID()}

	if got := finalized.Clients[0].CredentialRefs; !slices.Equal(got, declared) {
		t.Fatalf("defaulted client CredentialRefs = %v, want every declared slot", got)
	}

	if got := finalized.Clients[1].CredentialRefs; !slices.Equal(got, []integrationtypes.CredentialSlotID{testAuthCredentialRef.ID()}) {
		t.Fatalf("explicit client CredentialRefs = %v, want the declared slot kept", got)
	}

	if got := finalized.Connections[0].CredentialRefs; !slices.Equal(got, []integrationtypes.CredentialSlotID{testCredentialRef.ID()}) {
		t.Fatalf("defaulted connection CredentialRefs = %v, want the selecting slot", got)
	}

	if got := finalized.Connections[1].CredentialRefs; len(got) != 2 {
		t.Fatalf("explicit connection CredentialRefs = %v, want the authored slots kept", got)
	}

	if len(def.Clients[0].CredentialRefs) != 0 || len(def.Connections[0].CredentialRefs) != 0 {
		t.Fatal("expected finalize to leave the builder's registrations untouched")
	}

	names := lo.Map(finalized.Clients, func(client integrationtypes.ClientRegistration, _ int) string { return client.Ref.String() })
	if !slices.Equal(names, []string{clientRef.ID().String(), explicit.ID().String()}) {
		t.Fatalf("expected client order preserved, got %v", names)
	}
}
