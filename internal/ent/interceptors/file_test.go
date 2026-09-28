package interceptors

import (
	"context"
	"encoding/base64"
	"io"
	"testing"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql"
	"github.com/samber/mo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/ulids"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/theopenlane/core/common/storagetypes"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/intercept"
	"github.com/theopenlane/core/v2/internal/objects"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/eddy"
)

type downloadProvider struct {
	downloaded    []byte
	downloadCalls int
}

func (p *downloadProvider) Upload(context.Context, io.Reader, *storagetypes.UploadFileOptions) (*storagetypes.UploadedFileMetadata, error) {
	return &storagetypes.UploadedFileMetadata{}, nil
}

func (p *downloadProvider) Download(context.Context, *storagetypes.File, *storagetypes.DownloadFileOptions) (*storagetypes.DownloadedFileMetadata, error) {
	p.downloadCalls++
	return &storagetypes.DownloadedFileMetadata{File: p.downloaded}, nil
}

func (p *downloadProvider) Delete(context.Context, *storagetypes.File, *storagetypes.DeleteFileOptions) error {
	return nil
}

func (p *downloadProvider) GetPresignedURL(context.Context, *storagetypes.File, *storagetypes.PresignedURLOptions) (string, error) {
	return "", nil
}

func (p *downloadProvider) Exists(context.Context, *storagetypes.File) (bool, error) {
	return false, nil
}

func (p *downloadProvider) GetScheme() *string {
	return nil
}

func (p *downloadProvider) ListObjects(context.Context, string, int) ([]string, error) {
	return nil, nil
}

func (p *downloadProvider) ListBuckets() ([]string, error) {
	return nil, nil
}

func (p *downloadProvider) ProviderType() storagetypes.ProviderType {
	return storagetypes.S3Provider
}

func (p *downloadProvider) Close() error {
	return nil
}

func TestInterceptorPresignedURL_Base64Requested(t *testing.T) {
	provider := &downloadProvider{downloaded: []byte("downloaded-data")}
	objService := newObjectManager(t, provider)
	file := &generated.File{
		ID:                  "file-1",
		ProvidedFileName:    "avatar.png",
		DetectedContentType: "image/png",
		StorageProvider:     string(storagetypes.S3Provider),
		StorageVolume:       "bucket-1",
		StorageRegion:       "us-east-1",
		StoragePath:         "org-1/avatar.png",
	}

	query := &generated.FileQuery{}
	query.ObjectManager = objService

	intercepted := InterceptorPresignedURL().Intercept(generated.QuerierFunc(func(ctx context.Context, q generated.Query) (generated.Value, error) {
		return []*generated.File{file}, nil
	}))

	ctx := graphqlContextWithSelection("base64")
	result, err := intercepted.Query(ctx, query)
	require.NoError(t, err)

	files, ok := result.([]*generated.File)
	require.True(t, ok)
	require.Len(t, files, 1)

	assert.Equal(t, base64.StdEncoding.EncodeToString(provider.downloaded), files[0].Base64)
	assert.Equal(t, 1, provider.downloadCalls)
}

func TestInterceptorPresignedURL_Base64NotRequested(t *testing.T) {
	provider := &downloadProvider{downloaded: []byte("downloaded-data")}
	objService := newObjectManager(t, provider)
	file := &generated.File{
		ID:                  "file-1",
		ProvidedFileName:    "avatar.png",
		DetectedContentType: "image/png",
		StorageProvider:     string(storagetypes.S3Provider),
		StorageVolume:       "bucket-1",
		StorageRegion:       "us-east-1",
		StoragePath:         "org-1/avatar.png",
	}

	query := &generated.FileQuery{}
	query.ObjectManager = objService

	intercepted := InterceptorPresignedURL().Intercept(generated.QuerierFunc(func(ctx context.Context, q generated.Query) (generated.Value, error) {
		return []*generated.File{file}, nil
	}))

	ctx := graphqlContextWithSelection("id")
	result, err := intercepted.Query(ctx, query)
	require.NoError(t, err)

	files, ok := result.([]*generated.File)
	require.True(t, ok)
	require.Len(t, files, 1)

	assert.Empty(t, files[0].Base64)
	assert.Equal(t, 0, provider.downloadCalls)
}

func TestInterceptorPresignedURL_Base64DatabaseProvider(t *testing.T) {
	provider := &downloadProvider{downloaded: []byte("downloaded-data")}
	objService := newObjectManager(t, provider)
	file := &generated.File{
		ID:              "file-1",
		StorageProvider: string(storagetypes.DatabaseProvider),
		FileContents:    []byte("db-contents"),
	}

	query := &generated.FileQuery{}
	query.ObjectManager = objService

	intercepted := InterceptorPresignedURL().Intercept(generated.QuerierFunc(func(ctx context.Context, q generated.Query) (generated.Value, error) {
		return []*generated.File{file}, nil
	}))

	ctx := graphqlContextWithSelection("base64")
	result, err := intercepted.Query(ctx, query)
	require.NoError(t, err)

	files, ok := result.([]*generated.File)
	require.True(t, ok)
	require.Len(t, files, 1)

	assert.Equal(t, base64.StdEncoding.EncodeToString(file.FileContents), files[0].Base64)
	assert.Equal(t, 0, provider.downloadCalls)
}

func newObjectManager(t *testing.T, provider storage.Provider) *objects.Service {
	t.Helper()

	pool := eddy.NewClientPool[storage.Provider](time.Minute)
	clientService := eddy.NewClientService[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions](pool)

	builder := &eddy.BuilderFunc[storage.Provider, storage.ProviderCredentials, *storage.ProviderOptions]{
		Type: provider.ProviderType().String(),
		Func: func(context.Context, storage.ProviderCredentials, *storage.ProviderOptions) (storage.Provider, error) {
			return provider, nil
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

	return objects.NewService(objects.Config{
		Resolver:      resolver,
		ClientService: clientService,
	})
}

func graphqlContextWithSelection(field string) context.Context {
	opCtx := &graphql.OperationContext{
		Doc:       &ast.QueryDocument{},
		Variables: map[string]any{},
	}
	ctx := graphql.WithOperationContext(context.Background(), opCtx)

	return graphql.WithFieldContext(ctx, &graphql.FieldContext{
		Field: graphql.CollectedField{
			Selections: ast.SelectionSet{
				&ast.Field{Name: field},
			},
		},
	})
}

func TestInterceptorFileOrgFilter(t *testing.T) {
	orgID := ulids.New().String()

	testCases := []struct {
		name         string
		caller       *auth.Caller
		expectFilter bool
		expectOrgArg bool
	}{
		{
			name:         "caller with an org is filtered to its org",
			caller:       &auth.Caller{SubjectID: ulids.New().String(), OrganizationID: orgID, OrganizationIDs: []string{orgID}},
			expectFilter: true,
			expectOrgArg: true,
		},
		{
			name:         "caller without orgs is filtered to files without an org",
			caller:       &auth.Caller{SubjectID: ulids.New().String()},
			expectFilter: true,
		},
		{
			name:   "caller with org filter bypass is not filtered",
			caller: (&auth.Caller{SubjectID: ulids.New().String()}).WithCapabilities(auth.CapBypassOrgFilter),
		},
		{
			name:   "system admin is not filtered",
			caller: (&auth.Caller{SubjectID: ulids.New().String()}).WithCapabilities(auth.CapSystemAdmin),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			traverse, ok := InterceptorFile().(intercept.TraverseFunc)
			require.True(t, ok)

			query := &mockQuery{typ: generated.TypeFile}
			require.NoError(t, traverse(auth.WithCaller(context.Background(), tc.caller), query))

			if !tc.expectFilter {
				assert.Empty(t, query.predicates)
				return
			}

			require.NotEmpty(t, query.predicates)

			selector := sql.Select("*").From(sql.Table(file.Table))
			for _, p := range query.predicates {
				p(selector)
			}

			stmt, args := selector.Query()
			assert.Contains(t, stmt, file.FieldSystemOwned)

			if tc.expectOrgArg {
				assert.Contains(t, args, any(orgID))
			}
		})
	}
}
