package objstore

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/samber/mo"

	"github.com/theopenlane/eddy"
	"github.com/theopenlane/iam/auth"
	pkgobjects "github.com/theopenlane/shared/objects"
	"github.com/theopenlane/shared/objects/storage"
	storagetypes "github.com/theopenlane/shared/objects/storage/types"
	"github.com/theopenlane/utils/contextx"
)

type fakeProvider struct {
	id               string
	presignURL       string
	presignErr       error
	presignCallCount int
	uploadMetadata   *storagetypes.UploadedFileMetadata
	uploadErr        error
}

func (f *fakeProvider) Upload(context.Context, io.Reader, *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	if f.uploadErr != nil {
		return nil, f.uploadErr
	}
	if f.uploadMetadata != nil {
		return f.uploadMetadata, nil
	}
	return &storagetypes.UploadedFileMetadata{}, nil
}

func (f *fakeProvider) Download(context.Context, *storagetypes.File, *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	return nil, nil
}

func (f *fakeProvider) Delete(context.Context, *storagetypes.File, *storagetypes.DeleteFileOptions) error {
	return nil
}

func (f *fakeProvider) GetPresignedURL(context.Context, *storagetypes.File, *storagetypes.PresignedURLOptions) (string, error) {
	f.presignCallCount++
	if f.presignErr != nil {
		return "", f.presignErr
	}
	return f.presignURL, nil
}

func (f *fakeProvider) Exists(context.Context, *storagetypes.File) (bool, error) {
	return true, nil
}

func (f *fakeProvider) GetScheme() *string {
	return nil
}

func (f *fakeProvider) ListBuckets() ([]string, error) {
	return nil, nil
}

func (f *fakeProvider) ProviderType() storagetypes.ProviderType {
	return storagetypes.ProviderType(f.id)
}

func (f *fakeProvider) Close() error {
	return nil
}

func TestProviderCacheKeyString(t *testing.T) {
	key := ProviderCacheKey{TenantID: "tenant", IntegrationType: "s3"}
	if got := key.String(); got != "tenant:s3" {
		t.Fatalf("expected cache key to be %q, got %q", "tenant:s3", got)
	}
}

func TestServiceBuildResolutionContextAppliesHints(t *testing.T) {
	svc := &Service{}
	ctx := context.Background()

	opts := &storage.UploadOptions{
		FileMetadata: pkgobjects.FileMetadata{
			ProviderHints: &storage.ProviderHints{
				PreferredProvider: storage.S3Provider,
				Metadata: map[string]string{
					"size_bytes": "1024",
				},
			},
		},
	}

	ctx = svc.buildResolutionContext(ctx, opts)

	if pref, ok := contextx.From[PreferredProviderHint](ctx); !ok || storagetypes.ProviderType(pref) != storage.S3Provider {
		t.Fatal("expected preferred provider hint to be applied to context")
	}
	if size, ok := contextx.From[SizeBytesHint](ctx); !ok || int64(size) != 1024 {
		t.Fatalf("expected size hint to be applied")
	}
}

func TestServiceResolveUploadProviderSuccess(t *testing.T) {
	orgID := ulid.Make().String()
	ctx := auth.WithAuthenticatedUser(context.Background(), &auth.AuthenticatedUser{
		SubjectID:          ulid.Make().String(),
		OrganizationID:     orgID,
		AuthenticationType: auth.APITokenAuthentication,
	})

	pool := eddy.NewClientPool[storage.Provider](time.Minute)
	clientService := eddy.NewClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](pool)

	expectedProvider := &fakeProvider{id: "fake"}
	var buildCalls int

	builder := &eddy.BuilderFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Type: "fake",
		Func: func(context.Context, storage.ProviderCredentials, *storage.ProviderOptions) (storage.Provider, error) {
			buildCalls++
			return expectedProvider, nil
		},
	}

	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	resolver.AddRule(&eddy.RuleFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		EvaluateFunc: func(context.Context) mo.Option[eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]] {
			return mo.Some(eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
				Builder: builder,
				Output:  storage.ProviderCredentials{AccessKeyID: "id"},
				Config:  storage.NewProviderOptions(),
			})
		},
	})

	service := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
	})

	provider, err := service.resolveUploadProvider(ctx, &storage.UploadOptions{})
	if err != nil {
		t.Fatalf("expected no error resolving provider, got %v", err)
	}

	if provider != expectedProvider {
		t.Fatalf("expected provider %v, got %v", expectedProvider, provider)
	}

	// second call should use cache
	provider2, err := service.resolveUploadProvider(ctx, &storage.UploadOptions{})
	if err != nil {
		t.Fatalf("expected cached provider, got error %v", err)
	}

	if provider2 != expectedProvider {
		t.Fatal("expected cached provider to be returned")
	}

	if buildCalls != 1 {
		t.Fatalf("expected builder to be called once, got %d", buildCalls)
	}
}

func TestServiceResolveUploadProviderErrors(t *testing.T) {
	ctx := context.Background()
	clientService := eddy.NewClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](eddy.NewClientPool[storage.Provider](time.Minute))

	service := NewService(Config{
		Resolver:      eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](),
		ClientService: clientService,
	})

	if _, err := service.resolveUploadProvider(ctx, &storage.UploadOptions{}); !errors.Is(err, ErrProviderResolutionFailed) {
		t.Fatalf("expected ErrProviderResolutionFailed, got %v", err)
	}

	resolverMissingBuilder := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	resolverMissingBuilder.AddRule(&eddy.RuleFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		EvaluateFunc: func(context.Context) mo.Option[eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]] {
			return mo.Some(eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{})
		},
	})

	service.resolver = resolverMissingBuilder

	if _, err := service.resolveUploadProvider(ctx, &storage.UploadOptions{}); !errors.Is(err, ErrProviderResolutionFailed) {
		t.Fatalf("expected ErrProviderResolutionFailed when builder missing, got %v", err)
	}

	builder := &eddy.BuilderFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Type: "fake",
		Func: func(context.Context, storage.ProviderCredentials, *storage.ProviderOptions) (storage.Provider, error) {
			return &fakeProvider{id: "fake"}, nil
		},
	}

	resolverNoOrg := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	resolverNoOrg.AddRule(&eddy.RuleFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		EvaluateFunc: func(context.Context) mo.Option[eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]] {
			return mo.Some(eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
				Builder: builder,
				Output:  storage.ProviderCredentials{},
				Config:  storage.NewProviderOptions(),
			})
		},
	})

	service.resolver = resolverNoOrg

	if _, err := service.resolveUploadProvider(context.Background(), &storage.UploadOptions{}); !errors.Is(err, ErrNoOrganizationID) {
		t.Fatalf("expected ErrNoOrganizationID, got %v", err)
	}
}

func TestServiceResolveDownloadProvider(t *testing.T) {
	pool := eddy.NewClientPool[storage.Provider](time.Minute)
	clientService := eddy.NewClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](pool)

	expectedProvider := &fakeProvider{id: "fake"}
	builder := &eddy.BuilderFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Type: "fake",
		Func: func(context.Context, storage.ProviderCredentials, *storage.ProviderOptions) (storage.Provider, error) {
			return expectedProvider, nil
		},
	}

	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	resolver.AddRule(&eddy.RuleFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		EvaluateFunc: func(context.Context) mo.Option[eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]] {
			return mo.Some(eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
				Builder: builder,
				Output:  storage.ProviderCredentials{},
				Config:  storage.NewProviderOptions(storage.WithBucket("bucket")),
			})
		},
	})

	service := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
	})

	file := &storagetypes.File{
		ID: "file",
		FileMetadata: storagetypes.FileMetadata{
			ProviderType:  storagetypes.ProviderType(builder.Type),
			Bucket:        "bucket",
			Key:           "key",
			ProviderHints: &storagetypes.ProviderHints{},
		},
	}

	provider, err := service.resolveDownloadProvider(context.Background(), file)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if provider != expectedProvider {
		t.Fatal("expected resolved provider to match builder result")
	}
}

func TestServiceGetPresignedURL_ProviderMode(t *testing.T) {
	pool := eddy.NewClientPool[storage.Provider](time.Minute)
	clientService := eddy.NewClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](pool)

	trackedProvider := &fakeProvider{
		id:         string(storagetypes.S3Provider),
		presignURL: "provider-url",
	}

	builder := &eddy.BuilderFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Type: string(storagetypes.S3Provider),
		Func: func(context.Context, storage.ProviderCredentials, *storage.ProviderOptions) (storage.Provider, error) {
			return trackedProvider, nil
		},
	}

	resolver := eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]()
	resolver.AddRule(&eddy.RuleFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		EvaluateFunc: func(context.Context) mo.Option[eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]] {
			return mo.Some(eddy.Result[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
				Builder: builder,
				Output:  storage.ProviderCredentials{},
				Config:  storage.NewProviderOptions(),
			})
		},
	})

	service := NewService(Config{
		Resolver:      resolver,
		ClientService: clientService,
	})

	file := &storagetypes.File{
		ProviderType: storagetypes.S3Provider,
		FileMetadata: storagetypes.FileMetadata{
			ProviderType: storagetypes.S3Provider,
		},
	}

	url, err := service.GetPresignedURL(context.Background(), file, time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if url != trackedProvider.presignURL {
		t.Fatalf("expected provider URL, got %q", url)
	}

	if trackedProvider.presignCallCount != 1 {
		t.Fatalf("expected provider GetPresignedURL to be called once, got %d", trackedProvider.presignCallCount)
	}
}

func TestServiceResolveDownloadProviderNoResult(t *testing.T) {
	service := NewService(Config{
		Resolver:      eddy.NewResolver[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](),
		ClientService: eddy.NewClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](eddy.NewClientPool[storage.Provider](time.Minute)),
	})

	file := &storagetypes.File{ID: "file"}
	if _, err := service.resolveDownloadProvider(context.Background(), file); !errors.Is(err, ErrProviderResolutionFailed) {
		t.Fatalf("expected ErrProviderResolutionFailed, got %v", err)
	}
}
