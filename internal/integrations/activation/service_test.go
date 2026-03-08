package activation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations"
	"github.com/theopenlane/core/internal/integrations/types"
)

func TestNewServiceValidatesDependencies(t *testing.T) {
	t.Parallel()

	_, err := NewService(nil, &fakeOperationRunner{}, &fakePayloadMinter{})
	if !errors.Is(err, ErrStoreRequired) {
		t.Fatalf("expected ErrStoreRequired, got %v", err)
	}

	_, err = NewService(&fakeCredentialWriter{}, &fakeOperationRunner{}, nil)
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

	svc := mustNewService(t, writer, runner, minter)

	_, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:        "org-1",
		Provider:     provider,
		ProviderData: json.RawMessage(`{"key":"value"}`),
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
	minter := &fakePayloadMinter{
		returnPayload: &models.CredentialSet{OAuthAccessToken: "minted-token"},
	}

	svc := mustNewService(t, writer, runner, minter)

	result, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:        "org-1",
		Provider:     provider,
		ProviderData: json.RawMessage(`{"key":"value"}`),
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
	if writer.lastAuthKind != types.AuthKindOAuth2 {
		t.Fatalf("expected minted credential to be persisted as oauth2, got kind %s", writer.lastAuthKind)
	}
}

func TestConfigureNoValidateSkipsHealthCheck(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := types.ProviderType("acme")

	writer := &fakeCredentialWriter{}
	runner := &fakeOperationRunner{validateCalled: false}
	minter := &fakePayloadMinter{}

	svc := mustNewService(t, writer, runner, minter)

	_, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:        "org-1",
		Provider:     provider,
		ProviderData: json.RawMessage(`{"key":"value"}`),
		Validate:     false,
	})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if runner.validateCalled {
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

	svc := mustNewService(t, writer, runner, minter)

	providerData := json.RawMessage(`{"region":"us-east-1"}`)
	_, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:        "org-1",
		Provider:     provider,
		ProviderData: providerData,
		Validate:     true,
	})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if minter.lastRequest.Provider != provider {
		t.Fatalf("expected minter to receive provider %s, got %s", provider, minter.lastRequest.Provider)
	}
	if minter.lastRequest.OrgID != "org-1" {
		t.Fatalf("expected minter to receive org-1, got %s", minter.lastRequest.OrgID)
	}
	if string(minter.lastRequest.Credential.ProviderData) != string(providerData) {
		t.Fatalf(
			"expected minter to receive provider data %s, got %s",
			string(providerData),
			string(minter.lastRequest.Credential.ProviderData),
		)
	}
}

func TestConfigureOperationsRequiredWhenValidating(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := types.ProviderType("acme")

	writer := &fakeCredentialWriter{}
	minter := &fakePayloadMinter{}

	svc := mustNewService(t, writer, nil, minter)

	_, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:        "org-1",
		Provider:     provider,
		ProviderData: json.RawMessage(`{"token":"value"}`),
		Validate:     true,
	})
	if !errors.Is(err, ErrHealthValidatorRequired) {
		t.Fatalf("expected ErrHealthValidatorRequired, got %v", err)
	}
	if writer.saveCount != 0 {
		t.Fatalf("expected no saves when health validator not configured, got %d", writer.saveCount)
	}
}

func TestConfigureRequiresOrgID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	provider := types.ProviderType("acme")

	svc := mustNewService(t, &fakeCredentialWriter{}, &fakeOperationRunner{}, &fakePayloadMinter{})

	_, err := svc.Configure(ctx, ConfigureRequest{
		Provider: provider,
		Validate: true,
	})
	if !errors.Is(err, integrations.ErrOrgIDRequired) {
		t.Fatalf("expected ErrOrgIDRequired, got %v", err)
	}
}

func TestConfigureRequiresProvider(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	svc := mustNewService(t, &fakeCredentialWriter{}, &fakeOperationRunner{}, &fakePayloadMinter{})

	_, err := svc.Configure(ctx, ConfigureRequest{
		OrgID:    "org-1",
		Validate: true,
	})
	if !errors.Is(err, types.ErrProviderTypeRequired) {
		t.Fatalf("expected ErrProviderTypeRequired, got %v", err)
	}
}

// mustNewService constructs a Service for tests, panicking on error
func mustNewService(t *testing.T, writer CredentialWriter, runner HealthValidator, minter CredentialMinter) *Service {
	t.Helper()

	svc, err := NewService(writer, runner, minter)
	if err != nil {
		t.Fatalf("activation.NewService error: %v", err)
	}

	return svc
}

// fakeCredentialWriter records credential saves
type fakeCredentialWriter struct {
	saveCount    int
	lastPayload  models.CredentialSet
	lastAuthKind types.AuthKind
	saveErr      error
}

func (f *fakeCredentialWriter) SaveCredential(_ context.Context, _ string, _ types.ProviderType, authKind types.AuthKind, payload models.CredentialSet) (models.CredentialSet, error) {
	if f.saveErr != nil {
		return models.CredentialSet{}, f.saveErr
	}
	f.saveCount++
	f.lastAuthKind = authKind
	f.lastPayload = payload
	return payload, nil
}

// fakeOperationRunner records health validation calls.
type fakeOperationRunner struct {
	status         types.OperationStatus
	validateCalled bool
	runErr         error
}

func (f *fakeOperationRunner) ValidateProviderHealth(_ context.Context, _ string, _ types.ProviderType, _ models.CredentialSet) (types.OperationResult, error) {
	f.validateCalled = true
	if f.runErr != nil {
		return types.OperationResult{}, f.runErr
	}
	return types.OperationResult{Status: f.status}, nil
}

// fakePayloadMinter records mint calls and returns the subject credential unmodified
type fakePayloadMinter struct {
	lastRequest   types.CredentialMintRequest
	returnPayload *models.CredentialSet
	mintErr       error
}

func (f *fakePayloadMinter) MintCredential(_ context.Context, request types.CredentialMintRequest) (models.CredentialSet, error) {
	f.lastRequest = request
	if f.mintErr != nil {
		return models.CredentialSet{}, f.mintErr
	}
	if f.returnPayload != nil {
		return *f.returnPayload, nil
	}
	return request.Credential, nil
}
