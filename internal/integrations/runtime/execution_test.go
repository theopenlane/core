package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/iam/auth"
)

func TestExecuteOperationNilInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.ExecuteOperation(context.Background(), nil, types.OperationRegistration{}, nil, nil)
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

func TestExecuteOperationInvalidConfig(t *testing.T) {
	t.Parallel()

	reg := registry.New()
	rt := NewForTesting(reg)

	op := types.OperationRegistration{
		Name:         "test-op",
		ConfigSchema: json.RawMessage(`{"type":"object","required":["url"]}`),
	}

	_, err := rt.ExecuteOperation(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}, op, nil, json.RawMessage(`{}`))
	if !errors.Is(err, ErrOperationConfigInvalid) {
		t.Fatalf("expected ErrOperationConfigInvalid, got %v", err)
	}
}

func TestExecuteOperationValidConfigNoSchema(t *testing.T) {
	t.Parallel()

	reg := registry.New()

	called := false
	_ = reg.Register(types.Definition{
		DefinitionSpec: types.DefinitionSpec{
			ID: "test-def",
		},
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
	}, nil, nil)
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
	}, nil, nil)
	if !errors.Is(err, handlerErr) {
		t.Fatalf("expected handler error, got %v", err)
	}
}

func TestExecuteOperationValidConfigWithSchema(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())

	called := false
	op := types.OperationRegistration{
		Name:         "test-op",
		ConfigSchema: json.RawMessage(`{"type":"object","required":["url"]}`),
		Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
			called = true
			return json.RawMessage(`{"ok":true}`), nil
		},
	}

	result, err := rt.ExecuteOperation(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}, op, nil, json.RawMessage(`{"url":"https://example.com"}`))
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
		Name:         "test-op",
		ConfigSchema: json.RawMessage(`{"type":"object","required":["url"]}`),
		Handle: func(ctx context.Context, req types.OperationRequest) (json.RawMessage, error) {
			called = true
			return nil, nil
		},
	}

	// Empty config should skip schema validation even when schema is present
	_, err := rt.ExecuteOperation(context.Background(), &ent.Integration{
		ID:           "install-1",
		DefinitionID: "test-def",
	}, op, nil, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !called {
		t.Fatal("expected handler to be called with nil config")
	}
}

func TestExecuteOperationWithCredentials(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())

	ref := types.NewCredentialSlotID("api-key")
	credentials := types.CredentialBindings{
		{Ref: ref, Credential: types.CredentialSet{Data: json.RawMessage(`{"key":"secret"}`)}},
	}

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
	}, credentials, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(captured.Credentials) != 1 {
		t.Fatalf("expected 1 credential binding, got %d", len(captured.Credentials))
	}

	if captured.Credentials[0].Ref != ref {
		t.Fatalf("expected credential ref %v, got %v", ref, captured.Credentials[0].Ref)
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
	}, nil, nil)
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
	}, nil, nil)
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
	}, nil, config)
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

func TestEnsureCallerOrg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     func() context.Context
		orgID   string
		wantOrg string
		wantCap auth.Capability
	}{
		{
			name:  "sets org on caller without one",
			orgID: "org-123",
			ctx: func() context.Context {
				return auth.WithCaller(context.Background(), &auth.Caller{
					Capabilities: auth.CapBypassOrgFilter | auth.CapInternalOperation,
				})
			},
			wantOrg: "org-123",
			wantCap: auth.CapBypassOrgFilter | auth.CapInternalOperation,
		},
		{
			name:  "preserves existing org",
			orgID: "org-override",
			ctx: func() context.Context {
				return auth.WithCaller(context.Background(), &auth.Caller{
					OrganizationID: "org-original",
					Capabilities:   auth.CapInternalOperation,
				})
			},
			wantOrg: "org-original",
			wantCap: auth.CapInternalOperation,
		},
		{
			name:  "preserves org from OrganizationIDs",
			orgID: "org-override",
			ctx: func() context.Context {
				return auth.WithCaller(context.Background(), &auth.Caller{
					OrganizationIDs: []string{"org-from-ids"},
				})
			},
			wantOrg: "org-from-ids",
		},
		{
			name:  "no-op for empty orgID",
			orgID: "",
			ctx: func() context.Context {
				return auth.WithCaller(context.Background(), &auth.Caller{
					Capabilities: auth.CapInternalOperation,
				})
			},
		},
		{
			name:  "no-op when no caller in context",
			orgID: "org-123",
			ctx:   context.Background,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ensureCallerOrg(tc.ctx(), tc.orgID)
			caller, ok := auth.CallerFromContext(result)

			if tc.wantOrg == "" {
				if ok && caller != nil {
					if orgID, has := caller.ActiveOrg(); has && orgID != "" {
						assert.Assert(t, orgID == "", "expected no org, got %q", orgID)
					}
				}

				return
			}

			assert.Assert(t, ok && caller != nil, "expected caller in context")

			orgID, hasOrg := caller.ActiveOrg()
			assert.Assert(t, hasOrg, "expected caller to have active org")
			assert.Equal(t, tc.wantOrg, orgID)

			if tc.wantCap != 0 {
				assert.Assert(t, caller.Has(tc.wantCap), "expected capabilities to be preserved")
			}
		})
	}
}
