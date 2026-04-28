package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/samber/do/v2"
	"github.com/theopenlane/iam/auth"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestExecuteOperationNilInstallation(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())
	_, err := rt.ExecuteOperation(context.Background(), nil, types.OperationRegistration{}, nil, nil)
	if !errors.Is(err, ErrInstallationRequired) {
		t.Fatalf("expected ErrInstallationRequired, got %v", err)
	}
}

func TestBootstrapHandlerContextRuntimeRehydratesProcessDependencies(t *testing.T) {
	t.Parallel()

	db := &ent.Client{}
	injector := do.New()
	do.ProvideValue(injector, db)

	rt := &Runtime{injector: injector}
	caller := &auth.Caller{
		SubjectID:      "user_123",
		OrganizationID: "org_123",
	}
	metadata := types.ExecutionMetadata{
		OwnerID:      "org_123",
		DefinitionID: "email",
		Operation:    "send_email",
		Runtime:      true,
	}

	ctx := auth.WithCaller(context.Background(), caller)
	gotCtx, integration, gotMetadata, err := rt.bootstrapHandlerContext(ctx, metadata)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if integration != nil {
		t.Fatalf("expected no integration for runtime execution, got %#v", integration)
	}
	if gotMetadata != metadata {
		t.Fatalf("expected metadata passthrough, got %#v", gotMetadata)
	}
	if ent.FromContext(gotCtx) != db {
		t.Fatal("expected ent client to be reattached to handler context")
	}

	gotCaller, ok := auth.CallerFromContext(gotCtx)
	if !ok || gotCaller == nil {
		t.Fatal("expected original caller to remain on handler context")
	}
	if gotCaller.SubjectID != caller.SubjectID || gotCaller.OrganizationID != caller.OrganizationID {
		t.Fatalf("expected original caller, got %#v", gotCaller)
	}

	gotExecution, ok := types.ExecutionMetadataFromContext(gotCtx)
	if !ok {
		t.Fatal("expected execution metadata on handler context")
	}
	if gotExecution != metadata {
		t.Fatalf("expected execution metadata %#v, got %#v", metadata, gotExecution)
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
