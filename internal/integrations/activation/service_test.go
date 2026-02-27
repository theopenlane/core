package activation

import (
	"context"
	"errors"
	"testing"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/internal/keymaker"
	"github.com/theopenlane/core/internal/keystore"
)

func TestNewServiceValidatesDependencies(t *testing.T) {
	t.Parallel()

	resolver := &fakeProviderResolver{provider: &fakeProvider{providerType: types.ProviderType("acme")}}
	sessions := keymaker.NewMemorySessionStore()

	km, err := keymaker.NewService(resolver, &fakeCredentialWriter{}, sessions, keymaker.ServiceOptions{})
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	_, err = NewService(km, nil, &fakeOperationRunner{}, &fakePayloadMinter{})
	if !errors.Is(err, ErrStoreRequired) {
		t.Fatalf("expected ErrStoreRequired, got %v", err)
	}

	_, err = NewService(nil, &fakeCredentialWriter{}, &fakeOperationRunner{}, &fakePayloadMinter{})
	if !errors.Is(err, ErrKeymakerRequired) {
		t.Fatalf("expected ErrKeymakerRequired, got %v", err)
	}

	_, err = NewService(km, &fakeCredentialWriter{}, &fakeOperationRunner{}, nil)
	if !errors.Is(err, ErrMinterRequired) {
		t.Fatalf("expected ErrMinterRequired, got %v", err)
	}
}

func TestConfigureSkipsPersistOnHealthFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := types.ProviderType("acme")

	writer := &fakeCredentialWriter{}
	runner := &fakeOperationRunner{status: types.OperationStatusFailed}
	minter := &fakePayloadMinter{}

	svc := mustNewService(t, writer, runner, minter, provider)

	_, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:        "org-1",
		Provider:     provider,
		ProviderData: map[string]any{"key": "value"},
		Validate:     true,
	})
	if !errors.Is(err, ErrHealthCheckFailed) {
		t.Fatalf("expected ErrHealthCheckFailed, got %v", err)
	}
	if writer.saveCount != 0 {
		t.Fatalf("expected no credential saves when health check fails, got %d", writer.saveCount)
	}
}

func TestConfigurePersistsOnHealthSuccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := types.ProviderType("acme")

	writer := &fakeCredentialWriter{}
	runner := &fakeOperationRunner{status: types.OperationStatusOK}
	minter := &fakePayloadMinter{}

	svc := mustNewService(t, writer, runner, minter, provider)

	result, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:        "org-1",
		Provider:     provider,
		ProviderData: map[string]any{"key": "value"},
		Validate:     true,
	})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if writer.saveCount != 1 {
		t.Fatalf("expected one credential save, got %d", writer.saveCount)
	}
	if result.HealthResult == nil {
		t.Fatalf("expected health result in response")
	}
	if result.HealthResult.Status != types.OperationStatusOK {
		t.Fatalf("expected health status ok, got %s", result.HealthResult.Status)
	}
}

func TestConfigureNoValidateSkipsHealthCheck(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := types.ProviderType("acme")

	writer := &fakeCredentialWriter{}
	runner := &fakeOperationRunner{runCalled: false}
	minter := &fakePayloadMinter{}

	svc := mustNewService(t, writer, runner, minter, provider)

	_, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:        "org-1",
		Provider:     provider,
		ProviderData: map[string]any{"key": "value"},
		Validate:     false,
	})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if runner.runCalled || runner.runWithPayloadCalled {
		t.Fatalf("expected no operation runner calls when Validate=false")
	}
	if writer.saveCount != 1 {
		t.Fatalf("expected one credential save, got %d", writer.saveCount)
	}
}

func TestConfigureMintReceivesBuiltPayload(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := types.ProviderType("acme")

	writer := &fakeCredentialWriter{}
	runner := &fakeOperationRunner{status: types.OperationStatusOK}
	minter := &fakePayloadMinter{}

	svc := mustNewService(t, writer, runner, minter, provider)

	providerData := map[string]any{"serviceAccountKey": "test-key"}
	_, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:        "org-1",
		Provider:     provider,
		ProviderData: providerData,
		Validate:     true,
	})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if minter.lastSubject.Provider != provider {
		t.Fatalf("expected minter to receive provider %s, got %s", provider, minter.lastSubject.Provider)
	}
	if minter.lastSubject.OrgID != "org-1" {
		t.Fatalf("expected minter to receive org-1, got %s", minter.lastSubject.OrgID)
	}
	if minter.lastSubject.Credential.Data.ProviderData["serviceAccountKey"] != "test-key" {
		t.Fatalf("expected minter to receive provider data, got %v", minter.lastSubject.Credential.Data.ProviderData)
	}
}

func TestConfigureOperationsRequiredWhenValidating(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := types.ProviderType("acme")

	writer := &fakeCredentialWriter{}
	minter := &fakePayloadMinter{}

	svc := mustNewService(t, writer, nil, minter, provider)

	_, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:        "org-1",
		Provider:     provider,
		ProviderData: map[string]any{"token": "value"},
		Validate:     true,
	})
	if !errors.Is(err, ErrOperationsRequired) {
		t.Fatalf("expected ErrOperationsRequired, got %v", err)
	}
	if writer.saveCount != 0 {
		t.Fatalf("expected no saves when operations not configured, got %d", writer.saveCount)
	}
}

func TestConfigureRequiresOrgID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := types.ProviderType("acme")

	svc := mustNewService(t, &fakeCredentialWriter{}, &fakeOperationRunner{}, &fakePayloadMinter{}, provider)

	_, err := svc.Configure(ctx, ConfigureRequest{
		Provider: provider,
		Validate: true,
	})
	if !errors.Is(err, keystore.ErrOrgIDRequired) {
		t.Fatalf("expected ErrOrgIDRequired, got %v", err)
	}
}

func TestConfigureRequiresProvider(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := types.ProviderType("acme")

	svc := mustNewService(t, &fakeCredentialWriter{}, &fakeOperationRunner{}, &fakePayloadMinter{}, provider)

	_, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:    "org-1",
		Validate: true,
	})
	if !errors.Is(err, types.ErrProviderTypeRequired) {
		t.Fatalf("expected ErrProviderTypeRequired, got %v", err)
	}
}

// mustNewService constructs a Service for tests, panicking on error
func mustNewService(t *testing.T, writer CredentialWriter, runner OperationRunner, minter PayloadMinter, providerType types.ProviderType) *Service {
	t.Helper()

	resolver := &fakeProviderResolver{provider: &fakeProvider{providerType: providerType}}
	sessions := keymaker.NewMemorySessionStore()

	km, err := keymaker.NewService(resolver, writer, sessions, keymaker.ServiceOptions{})
	if err != nil {
		t.Fatalf("keymaker.NewService error: %v", err)
	}

	svc, err := NewService(km, writer, runner, minter)
	if err != nil {
		t.Fatalf("activation.NewService error: %v", err)
	}

	return svc
}

// fakeCredentialWriter records credential saves
type fakeCredentialWriter struct {
	saveCount int
	saveErr   error
}

func (f *fakeCredentialWriter) SaveCredential(_ context.Context, _ string, payload types.CredentialPayload) (types.CredentialPayload, error) {
	if f.saveErr != nil {
		return types.CredentialPayload{}, f.saveErr
	}
	f.saveCount++
	return payload, nil
}

// fakeOperationRunner records operation calls
type fakeOperationRunner struct {
	status               types.OperationStatus
	runCalled            bool
	runWithPayloadCalled bool
	runErr               error
}

func (f *fakeOperationRunner) Run(_ context.Context, _ types.OperationRequest) (types.OperationResult, error) {
	f.runCalled = true
	if f.runErr != nil {
		return types.OperationResult{}, f.runErr
	}
	return types.OperationResult{Status: f.status}, nil
}

func (f *fakeOperationRunner) RunWithPayload(_ context.Context, _ types.OperationRequest, _ types.CredentialPayload) (types.OperationResult, error) {
	f.runWithPayloadCalled = true
	if f.runErr != nil {
		return types.OperationResult{}, f.runErr
	}
	return types.OperationResult{Status: f.status}, nil
}

// fakePayloadMinter records mint calls and returns the subject credential unmodified
type fakePayloadMinter struct {
	lastSubject types.CredentialSubject
	mintErr     error
}

func (f *fakePayloadMinter) MintPayload(_ context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	f.lastSubject = subject
	if f.mintErr != nil {
		return types.CredentialPayload{}, f.mintErr
	}
	return subject.Credential, nil
}

// fakeProvider satisfies types.Provider for test keymaker construction
type fakeProvider struct {
	providerType types.ProviderType
}

func (p *fakeProvider) Type() types.ProviderType { return p.providerType }

func (p *fakeProvider) Capabilities() types.ProviderCapabilities {
	return types.ProviderCapabilities{}
}

func (p *fakeProvider) BeginAuth(_ context.Context, _ types.AuthContext) (types.AuthSession, error) {
	return nil, errors.New("not implemented")
}

func (p *fakeProvider) Mint(_ context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	return subject.Credential, nil
}

// fakeProviderResolver satisfies keymaker.ProviderResolver
type fakeProviderResolver struct {
	provider types.Provider
}

func (r *fakeProviderResolver) Provider(_ types.ProviderType) (types.Provider, bool) {
	if r.provider == nil {
		return nil, false
	}
	return r.provider, true
}
