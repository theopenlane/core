package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/samber/do/v2"
	"github.com/theopenlane/iam/auth"
	"gotest.tools/v3/assert"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

// errTestCycle is the provider failure used by cycle classification tests
var errTestCycle = errors.New("provider request failed")

// reconcileEnvelope builds an installation-bound envelope for classification tests
func reconcileEnvelope(t *testing.T, integrationID string) operations.ReconcileEnvelope {
	t.Helper()

	oc := types.NewOperationContext("org-1", "DirectorySync", types.IntegrationSource{
		IntegrationID: integrationID,
		DefinitionID:  "test-def",
	})

	return operations.ReconcileEnvelope{OperationContext: oc}
}

func TestRegisterListenersRequiresGala(t *testing.T) {
	t.Parallel()

	err := NewForTesting(registry.New()).registerListeners()
	if !errors.Is(err, operations.ErrGalaRequired) {
		t.Fatalf("expected ErrGalaRequired, got %v", err)
	}
}

func TestOperationCancel(t *testing.T) {
	t.Parallel()

	envelope := operations.Envelope{OperationContext: types.NewOperationContext("org-1", "SeedDemoData", types.IntegrationSource{DefinitionID: "test-def", Runtime: true})}

	if !operationCancel(types.ExecutionPolicy{NoRetry: true})(context.Background(), envelope, errTestCycle) {
		t.Fatalf("expected a no-retry operation to cancel on any error")
	}

	if operationCancel(types.ExecutionPolicy{})(context.Background(), envelope, errTestCycle) {
		t.Fatalf("expected a retryable operation to retry a provider error")
	}

	if !operationCancel(types.ExecutionPolicy{})(context.Background(), envelope, &ent.NotFoundError{}) {
		t.Fatalf("expected a missing integration to cancel")
	}

	if !operationCancel(types.ExecutionPolicy{})(context.Background(), envelope, fmt.Errorf("send: %w", types.Terminal(errTestCycle, "recipient rejected"))) {
		t.Fatalf("expected a terminal failure to cancel")
	}

	if operationCancel(types.ExecutionPolicy{})(context.Background(), envelope, types.Degraded(errTestCycle, "scope missing")) {
		t.Fatalf("expected a degraded failure to retry")
	}
}

func TestCancelsRetryMatchesOperationCancel(t *testing.T) {
	t.Parallel()

	envelope := operations.Envelope{OperationContext: types.NewOperationContext("org-1", "SeedDemoData", types.IntegrationSource{DefinitionID: "test-def", Runtime: true})}

	tests := []struct {
		name   string
		policy types.ExecutionPolicy
		err    error
		want   bool
	}{
		{name: "provider error retries", err: errTestCycle},
		{name: "no-retry policy cancels", policy: types.ExecutionPolicy{NoRetry: true}, err: errTestCycle, want: true},
		{name: "missing integration cancels", err: &ent.NotFoundError{}, want: true},
		{name: "terminal failure cancels", err: fmt.Errorf("send: %w", types.Terminal(errTestCycle, "recipient rejected")), want: true},
		{name: "unhealthy failure retries", err: types.Unhealthy(errTestCycle, "needs reauthorization")},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := cancelsRetry(tc.policy, tc.err); got != tc.want {
				t.Fatalf("cancelsRetry() = %v, want %v", got, tc.want)
			}

			if got := operationCancel(tc.policy)(context.Background(), envelope, tc.err); got != tc.want {
				t.Fatalf("operationCancel() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestReconcileShouldCancelClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "plain error keeps the loop running",
			err:  errTestCycle,
			want: false,
		},
		{
			name: "operation disabled stops the loop",
			err:  fmt.Errorf("cycle: %w", operations.ErrOperationDisabled),
			want: true,
		},
		{
			name: "unhealthy failure stops the loop",
			err:  fmt.Errorf("cycle: %w", types.Unhealthy(errTestCycle, "needs reauthorization")),
			want: true,
		},
		{
			name: "degraded failure stops the loop",
			err:  fmt.Errorf("cycle: %w", types.Degraded(errTestCycle, "missing operation permission")),
			want: true,
		},
		{
			name: "definition no longer registered stops the loop",
			err:  fmt.Errorf("cycle: %w", registry.ErrDefinitionNotFound),
			want: true,
		},
		{
			name: "operation no longer registered stops the loop",
			err:  fmt.Errorf("cycle: %w", registry.ErrOperationNotFound),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := reconcileShouldCancel(t.Context(), registry.New(), reconcileEnvelope(t, "install-1"), tt.err)
			if got != tt.want {
				t.Fatalf("reconcileShouldCancel = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReconcileShouldCancelNotFoundScoping(t *testing.T) {
	t.Parallel()

	notFound := &ent.NotFoundError{}

	if !reconcileShouldCancel(t.Context(), registry.New(), reconcileEnvelope(t, "install-1"), notFound) {
		t.Fatal("expected not-found to stop an installation-bound loop")
	}

	if reconcileShouldCancel(t.Context(), registry.New(), reconcileEnvelope(t, ""), notFound) {
		t.Fatal("expected not-found to keep a runtime-bound sweep running")
	}
}

func TestOperationListenerInvokesExhaustedHook(t *testing.T) {
	t.Parallel()

	const (
		definitionID  = "exhaust-def"
		operationName = "exhaust-op"
	)

	reg := registry.New()
	assert.NilError(t, reg.Register(types.Definition{
		ID: definitionID,
		Operations: []types.OperationRegistration{{
			Name:   operationName,
			Handle: hookTestHandler(nil, errTestOperationFailed),
		}},
	}))

	listeners := reg.Listeners()
	assert.Equal(t, len(listeners), 1)

	g, err := gala.NewGala(context.Background(), gala.Config{DispatchMode: gala.DispatchModeInMemory})
	assert.NilError(t, err)

	rt := NewForTesting(reg)
	do.OverrideValue(rt.injector, g)

	var (
		gotEnvelope operations.Envelope
		gotErr      error
	)

	rt.SetExhaustedHook(func(_ context.Context, envelope operations.Envelope, err error) {
		gotEnvelope = envelope
		gotErr = err
	})

	assert.NilError(t, rt.registerListeners())

	envelope := operations.Envelope{
		OperationContext: types.NewOperationContext("org-1", operationName, types.IntegrationSource{DefinitionID: definitionID, Runtime: true}),
		Config:           json.RawMessage(`{}`),
	}

	_, err = g.EmitWithHeaders(auth.WithCaller(context.Background(), &auth.Caller{SubjectID: "test_subject"}), listeners[0].Topic, envelope, gala.Headers{})
	assert.NilError(t, err)
	assert.NilError(t, g.WaitIdle(context.Background()))

	assert.Equal(t, gotEnvelope.Operation, operationName)
	assert.Assert(t, errors.Is(gotErr, errTestOperationFailed), "got %v", gotErr)
}
