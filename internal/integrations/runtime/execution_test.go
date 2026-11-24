package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/openapi"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/integrations/operations"
	"github.com/theopenlane/core/v2/internal/integrations/registry"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
)

func TestExecuteOperationNilInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.ExecuteOperation(context.Background(), nil, types.OperationRegistration{}, nil)
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

func TestExecuteOperationInvalidConfig(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	rt := NewForTesting(reg)

	op := types.OperationRegistration{
		Name:  "test-op",
		Input: types.InputRegistration{Schema: json.RawMessage(`{"type":"object","required":["url"]}`)},
	}

	_, err := rt.ExecuteOperation(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}, op, json.RawMessage(`{}`))
	if !errors.Is(err, types.ErrOperationConfigInvalid) {
		t.Fatalf("expected ErrOperationConfigInvalid, got %v", err)
	}
}

func TestExecuteOperationValidConfigNoSchema(t *testing.T) {
	t.Parallel()

	reg := registry.New()

	called := false
	_ = reg.Register(types.Definition{
		ID: "test-def",
		Operations: []types.OperationRegistration{
			{
				Name: "test-op",
				Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
					called = true
					return json.RawMessage(`{"ok":true}`), nil
				},
			},
		},
	})

	rt := NewForTesting(reg)

	result, err := rt.ExecuteOperation(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}, types.OperationRegistration{
		Name: "test-op",
		Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
			called = true
			return json.RawMessage(`{"ok":true}`), nil
		},
	}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !called {
		t.Fatal("expected handler to be called")
	}

	if string(result) != `{"ok":true}` {
		t.Fatalf("expected {\"ok\":true}, got %s", string(result))
	}
}

func TestExecuteOperationHandlerError(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	handlerErr := errors.New("handler failed")

	_, err := rt.ExecuteOperation(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}, types.OperationRegistration{
		Name: "test-op",
		Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
			return nil, handlerErr
		},
	}, nil)
	if !errors.Is(err, handlerErr) {
		t.Fatalf("expected handler error, got %v", err)
	}
}

func TestExecuteOperationValidConfigWithSchema(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())

	called := false
	op := types.OperationRegistration{
		Name:  "test-op",
		Input: types.InputRegistration{Schema: json.RawMessage(`{"type":"object","required":["url"]}`)},
		Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
			called = true
			return json.RawMessage(`{"ok":true}`), nil
		},
	}

	result, err := rt.ExecuteOperation(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}, op, json.RawMessage(`{"url":"https://example.com"}`))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !called {
		t.Fatal("expected handler to be called")
	}

	if string(result) != `{"ok":true}` {
		t.Fatalf("expected ok response, got %s", string(result))
	}
}

func TestExecuteOperationEmptyConfigSkipsValidation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())

	called := false
	op := types.OperationRegistration{
		Name:  "test-op",
		Input: types.InputRegistration{Schema: json.RawMessage(`{"type":"object","required":["url"]}`)},
		Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
			called = true
			return nil, nil
		},
	}

	_, err := rt.ExecuteOperation(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}, op, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !called {
		t.Fatal("expected handler to be called with nil config")
	}
}

func TestExecuteOperationAppliesDefaultLookbackWhenNoLastRun(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())

	before := time.Now().UTC()

	var captured types.OperationRequest
	_, err := rt.ExecuteOperation(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}, types.OperationRegistration{
		Name: "test-op",
		Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
			captured = req
			return nil, nil
		},
	}, nil)
	assert.NilError(t, err)
	assert.Assert(t, captured.LastRunAt != nil, "expected LastRunAt to be set via default lookback")

	after := time.Now().UTC()
	expectedEarliest := before.Add(-defaultLookbackDuration)
	expectedLatest := after.Add(-defaultLookbackDuration)
	assert.Assert(t, !captured.LastRunAt.Before(expectedEarliest), "LastRunAt %v is before expected window start %v", captured.LastRunAt, expectedEarliest)
	assert.Assert(t, !captured.LastRunAt.After(expectedLatest), "LastRunAt %v is after expected window end %v", captured.LastRunAt, expectedLatest)
}

func TestExecuteOperationSkipDefaultLookbackLeavesLastRunAtNil(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())

	var captured types.OperationRequest
	_, err := rt.ExecuteOperation(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}, types.OperationRegistration{
		Name:                "test-op",
		SkipDefaultLookback: true,
		Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
			captured = req
			return nil, nil
		},
	}, nil)
	assert.NilError(t, err)
	assert.Assert(t, captured.LastRunAt == nil, "expected LastRunAt to be nil when SkipDefaultLookback is set")
}

func TestExecuteOperationPassesRequestFields(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	installation := &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}
	config := json.RawMessage(`{"key":"value"}`)

	var captured types.OperationRequest
	_, err := rt.ExecuteOperation(context.Background(), installation, types.OperationRegistration{
		Name: "test-op",
		Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
			captured = req
			return nil, nil
		},
	}, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if captured.Integration != installation {
		t.Fatal("expected installation to be passed through")
	}

	if string(captured.Config) != `{"key":"value"}` {
		t.Fatalf("expected config to be passed through, got %s", string(captured.Config))
	}
}

func TestExecuteOperationResolvesConfigFromInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	installation := &ent.Integration{
		ID:              "install-1",
		DefinitionID:    "test-def",
		OperationConfig: openapi.IntegrationOperationConfig{Operations: map[string]json.RawMessage{"test-op": json.RawMessage(`{"filterExpr":"payload.active"}`)}},
	}

	var captured types.OperationRequest
	_, err := rt.ExecuteOperation(context.Background(), installation, types.OperationRegistration{
		Name: "test-op",
		Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
			captured = req
			return nil, nil
		},
	}, nil)
	assert.NilError(t, err)
	assert.Equal(t, string(captured.Config), `{"filterExpr":"payload.active"}`)
}

func TestExecuteOperationRefusesDisabledOperation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	installation := &ent.Integration{
		ID:              "install-1",
		DefinitionID:    "test-def",
		OperationConfig: openapi.IntegrationOperationConfig{Operations: map[string]json.RawMessage{"test-op": json.RawMessage(`{"disable":true}`)}},
	}

	cases := map[string]types.OperationRegistration{
		"disabled for all": {
			Name:           "other-op",
			DisabledForAll: true,
		},
		"disabled for the installation": {
			Name: "test-op",
		},
	}

	for name, op := range cases {
		called := false
		op.Handle = func(context.Context, types.OperationRequest) (json.RawMessage, error) {
			called = true
			return nil, nil
		}

		_, err := rt.ExecuteOperation(context.Background(), installation, op, nil)
		if !errors.Is(err, operations.ErrOperationDisabled) {
			t.Fatalf("%s: expected ErrOperationDisabled, got %v", name, err)
		}

		if called {
			t.Fatalf("%s: expected the handler not to run", name)
		}
	}
}

func TestExecuteOperationExplicitConfigWinsOverStored(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	installation := &ent.Integration{
		ID:              "install-1",
		DefinitionID:    "test-def",
		OperationConfig: openapi.IntegrationOperationConfig{Operations: map[string]json.RawMessage{"test-op": json.RawMessage(`{"filterExpr":"payload.active"}`)}},
	}

	var captured types.OperationRequest
	_, err := rt.ExecuteOperation(context.Background(), installation, types.OperationRegistration{
		Name: "test-op",
		Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
			captured = req
			return nil, nil
		},
	}, json.RawMessage(`{"key":"value"}`))
	assert.NilError(t, err)
	assert.Equal(t, string(captured.Config), `{"key":"value"}`)
}

func TestExecuteRuntimeOperationKeepsDispatchCaller(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		caller *auth.Caller
	}{
		{name: "user caller", caller: &auth.Caller{SubjectID: "user_1", OrganizationID: "org_1"}},
		{name: "no caller"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := gala.WithOperationContext(context.Background(), gala.OperationContext{OwnerID: "org_1"})
			if tc.caller != nil {
				ctx = auth.WithCaller(ctx, tc.caller)
			}

			var (
				captured *auth.Caller
				found    bool
			)

			_, _, err := NewForTesting(registry.New()).executeResolvedOperation(ctx, nil, types.OperationRegistration{
				Name: "test-op",
				Handle: func(ctx context.Context, _ types.OperationRequest) (json.RawMessage, error) {
					captured, found = auth.CallerFromContext(ctx)
					return nil, nil
				},
			}, nil, false, operations.IngestOptions{})
			assert.NilError(t, err)

			if tc.caller == nil {
				assert.Check(t, !found, "a non-ingest runtime operation must not gain an integration caller")
				return
			}

			assert.Assert(t, found)
			assert.Equal(t, tc.caller.SubjectID, captured.SubjectID)
			assert.Check(t, !captured.Has(auth.CapBypassAuditLog))
			assert.Check(t, !captured.Has(auth.CapBypassOrgFilter))
		})
	}
}

// errTestHookFailed is the post-execution hook failure the runtime must log and not return
var errTestHookFailed = errors.New("post execution hook failed")

// errTestOperationFailed is the operation failure the post-execution hook tests report
var errTestOperationFailed = errors.New("operation failed")

// errTestNotFound is the missing-integration failure that cancels retries
var errTestNotFound = &ent.NotFoundError{}

// hookTestHandler returns an operation handler responding with response and err
func hookTestHandler(response json.RawMessage, err error) types.OperationHandler {
	return func(context.Context, types.OperationRequest) (json.RawMessage, error) {
		return response, err
	}
}

func TestHandleOperationPostExecutionHookOutcome(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		handle       types.OperationHandler
		policy       types.ExecutionPolicy
		wantErr      error
		wantFinal    bool
		wantResponse string
	}{
		{
			name:         "success is final and carries the response",
			handle:       hookTestHandler(json.RawMessage(`{"ok":true}`), nil),
			wantFinal:    true,
			wantResponse: `{"ok":true}`,
		},
		{
			name:    "retryable failure is not final",
			handle:  hookTestHandler(nil, errTestOperationFailed),
			wantErr: errTestOperationFailed,
		},
		{
			name:      "failure of a no-retry operation is final",
			handle:    hookTestHandler(nil, errTestOperationFailed),
			policy:    types.ExecutionPolicy{NoRetry: true},
			wantErr:   errTestOperationFailed,
			wantFinal: true,
		},
		{
			name:      "not-found failure is final",
			handle:    hookTestHandler(nil, errTestNotFound),
			wantErr:   errTestNotFound,
			wantFinal: true,
		},
		{
			name:      "terminal failure is final",
			handle:    hookTestHandler(nil, types.Terminal(errTestOperationFailed, "recipient rejected")),
			wantErr:   errTestOperationFailed,
			wantFinal: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reg := registry.New()
			assert.NilError(t, reg.Register(types.Definition{
				ID:         "hook-def",
				Operations: []types.OperationRegistration{{Name: "hook-op", Handle: tc.handle}},
			}))

			rt := NewForTesting(reg)

			var outcomes []OperationOutcome

			rt.SetPostExecutionHook(func(_ context.Context, outcome OperationOutcome) error {
				outcomes = append(outcomes, outcome)

				return errTestHookFailed
			})

			envelope := operations.Envelope{
				OperationContext: types.NewOperationContext("org-1", "hook-op", types.IntegrationSource{DefinitionID: "hook-def", Runtime: true}),
				Config:           json.RawMessage(`{}`),
			}

			err := rt.HandleOperation(context.Background(), envelope, tc.policy)

			switch tc.wantErr {
			case nil:
				assert.NilError(t, err)
			default:
				assert.Assert(t, errors.Is(err, tc.wantErr), "got %v", err)
				assert.Assert(t, !errors.Is(err, errTestHookFailed), "the hook failure must not be returned")
			}

			assert.Assert(t, len(outcomes) == 1, "expected one hook call, got %d", len(outcomes))
			assert.Equal(t, outcomes[0].Final, tc.wantFinal)
			assert.Equal(t, string(outcomes[0].Response), tc.wantResponse)
			assert.Equal(t, outcomes[0].Envelope.Operation, "hook-op")
			assert.Assert(t, errors.Is(outcomes[0].Err, tc.wantErr), "outcome error = %v", outcomes[0].Err)
		})
	}
}

func TestExecutionRunResultContextErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus enums.IntegrationRunStatus
		wantError  string
	}{
		{name: "canceled", err: fmt.Errorf("aikido: unable to export issues: %w", context.Canceled), wantStatus: enums.IntegrationRunStatusCancelled, wantError: runInterruptedError},
		{name: "deadline", err: context.DeadlineExceeded, wantStatus: enums.IntegrationRunStatusFailed, wantError: runTimedOutError},
		{name: "other", err: errors.New("boom"), wantStatus: enums.IntegrationRunStatusFailed, wantError: "boom"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := executionRunResult(true, nil, operations.IngestResult{}, tc.err)

			assert.Check(t, cmp.Equal(tc.wantStatus, result.Status))
			assert.Check(t, cmp.Equal(tc.wantError, result.Error))
		})
	}
}
