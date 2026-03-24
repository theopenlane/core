package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/integrations/registry"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestMergeCredentialsBothEmpty(t *testing.T) {
	t.Parallel()

	result := mergeCredentials(nil, nil)
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}
}

func TestMergeCredentialsCurrentOnly(t *testing.T) {
	t.Parallel()

	ref := types.NewCredentialSlotID("oauth")
	current := types.CredentialBindings{
		{Ref: ref, Credential: types.CredentialSet{Data: json.RawMessage(`{"token":"a"}`)}},
	}

	result := mergeCredentials(current, nil)
	if len(result) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(result))
	}

	if result[0].Ref != ref {
		t.Fatalf("expected ref %v, got %v", ref, result[0].Ref)
	}
}

func TestMergeCredentialsOverridesOnly(t *testing.T) {
	t.Parallel()

	ref := types.NewCredentialSlotID("api-key")
	overrides := types.CredentialBindings{
		{Ref: ref, Credential: types.CredentialSet{Data: json.RawMessage(`{"key":"b"}`)}},
	}

	result := mergeCredentials(nil, overrides)
	if len(result) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(result))
	}

	if string(result[0].Credential.Data) != `{"key":"b"}` {
		t.Fatalf("expected override credential, got %s", string(result[0].Credential.Data))
	}
}

func TestMergeCredentialsReplacesMatching(t *testing.T) {
	t.Parallel()

	ref := types.NewCredentialSlotID("oauth")
	current := types.CredentialBindings{
		{Ref: ref, Credential: types.CredentialSet{Data: json.RawMessage(`{"token":"old"}`)}},
	}
	overrides := types.CredentialBindings{
		{Ref: ref, Credential: types.CredentialSet{Data: json.RawMessage(`{"token":"new"}`)}},
	}

	result := mergeCredentials(current, overrides)
	if len(result) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(result))
	}

	if string(result[0].Credential.Data) != `{"token":"new"}` {
		t.Fatalf("expected new credential, got %s", string(result[0].Credential.Data))
	}
}

func TestMergeCredentialsAppendsNovelOverride(t *testing.T) {
	t.Parallel()

	refA := types.NewCredentialSlotID("oauth")
	refB := types.NewCredentialSlotID("api-key")
	current := types.CredentialBindings{
		{Ref: refA, Credential: types.CredentialSet{Data: json.RawMessage(`{"token":"a"}`)}},
	}
	overrides := types.CredentialBindings{
		{Ref: refB, Credential: types.CredentialSet{Data: json.RawMessage(`{"key":"b"}`)}},
	}

	result := mergeCredentials(current, overrides)
	if len(result) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(result))
	}

	if result[0].Ref != refA {
		t.Fatalf("expected first ref %v, got %v", refA, result[0].Ref)
	}

	if result[1].Ref != refB {
		t.Fatalf("expected second ref %v, got %v", refB, result[1].Ref)
	}
}

func TestMergeCredentialsMixedReplaceAndAppend(t *testing.T) {
	t.Parallel()

	refA := types.NewCredentialSlotID("oauth")
	refB := types.NewCredentialSlotID("api-key")
	refC := types.NewCredentialSlotID("service-account")
	current := types.CredentialBindings{
		{Ref: refA, Credential: types.CredentialSet{Data: json.RawMessage(`{"a":"old"}`)}},
		{Ref: refB, Credential: types.CredentialSet{Data: json.RawMessage(`{"b":"keep"}`)}},
	}
	overrides := types.CredentialBindings{
		{Ref: refA, Credential: types.CredentialSet{Data: json.RawMessage(`{"a":"new"}`)}},
		{Ref: refC, Credential: types.CredentialSet{Data: json.RawMessage(`{"c":"added"}`)}},
	}

	result := mergeCredentials(current, overrides)
	if len(result) != 3 {
		t.Fatalf("expected 3 bindings, got %d", len(result))
	}

	if string(result[0].Credential.Data) != `{"a":"new"}` {
		t.Fatalf("expected replaced credential for refA, got %s", string(result[0].Credential.Data))
	}

	if string(result[1].Credential.Data) != `{"b":"keep"}` {
		t.Fatalf("expected untouched credential for refB, got %s", string(result[1].Credential.Data))
	}

	if string(result[2].Credential.Data) != `{"c":"added"}` {
		t.Fatalf("expected appended credential for refC, got %s", string(result[2].Credential.Data))
	}
}

func TestSingleCredentialNoRefs(t *testing.T) {
	t.Parallel()

	result := singleCredential(nil, nil)
	if len(result.Data) != 0 {
		t.Fatalf("expected empty credential, got %v", result)
	}
}

func TestSingleCredentialMultipleRefs(t *testing.T) {
	t.Parallel()

	refA := types.NewCredentialSlotID("a")
	refB := types.NewCredentialSlotID("b")
	bindings := types.CredentialBindings{
		{Ref: refA, Credential: types.CredentialSet{Data: json.RawMessage(`{"a":1}`)}},
		{Ref: refB, Credential: types.CredentialSet{Data: json.RawMessage(`{"b":2}`)}},
	}

	result := singleCredential(bindings, []types.CredentialSlotID{refA, refB})
	if len(result.Data) != 0 {
		t.Fatalf("expected empty credential for multi-ref, got %v", result)
	}
}

func TestSingleCredentialExactlyOneRef(t *testing.T) {
	t.Parallel()

	ref := types.NewCredentialSlotID("oauth")
	bindings := types.CredentialBindings{
		{Ref: ref, Credential: types.CredentialSet{Data: json.RawMessage(`{"token":"x"}`)}},
	}

	result := singleCredential(bindings, []types.CredentialSlotID{ref})
	if string(result.Data) != `{"token":"x"}` {
		t.Fatalf("expected resolved credential, got %s", string(result.Data))
	}
}

func TestSingleCredentialRefNotFound(t *testing.T) {
	t.Parallel()

	ref := types.NewCredentialSlotID("missing")
	result := singleCredential(nil, []types.CredentialSlotID{ref})
	if len(result.Data) != 0 {
		t.Fatalf("expected empty credential for missing ref, got %v", result)
	}
}

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

func TestExecuteOperationWithCredentialOverrides(t *testing.T) {
	t.Parallel()

	rt := NewForTesting(registry.New())

	ref := types.NewCredentialSlotID("api-key")
	overrides := types.CredentialBindings{
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
	}, overrides, nil)
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
