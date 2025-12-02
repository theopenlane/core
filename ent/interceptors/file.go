package interceptors

import (
	"context"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/shared/logx"
	"github.com/theopenlane/shared/objects/storage/proxy"
	storagetypes "github.com/theopenlane/shared/objects/storage/types"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/file"
	"github.com/theopenlane/ent/generated/intercept"
	"github.com/theopenlane/ent/generated/organization"
)

// InterceptorFile is an ent interceptor that filters the file query on the organization id
// this is slightly different from the organization interceptor because this is formatted differently
// then other schemas and is not always required so keeping it separate
func InterceptorFile() ent.Interceptor {
	return intercept.TraverseFunc(func(ctx context.Context, q intercept.Query) error {
		logx.FromContext(ctx).Debug().Msg("InterceptorFile")

		var orgs []string
		if anon, ok := auth.AnonymousTrustCenterUserFromContext(ctx); ok {
			// q.WhereP(trustcenter.IDEQ(anon.TrustCenterID))
			orgs = []string{anon.OrganizationID}
		} else {
			au, err := auth.GetAuthenticatedUserFromContext(ctx)
			if err != nil {
				return err
			}

			if au.IsSystemAdmin {
				logx.FromContext(ctx).Debug().Msg("user is system admin, skipping organization filter")

				return nil
			}

			orgs = au.OrganizationIDs
		}

		if len(orgs) == 0 {
			return nil
		}

		// filter on the organization id or empty organization id
		// which would happen on something like a user file
		q.WhereP(
			file.Or(
				file.HasOrganizationWith(organization.IDIn(orgs...)),
				file.Not(file.HasOrganization()),
			),
		)

		return nil
	})
}

// InterceptorPresignedURL is an ent interceptor that sets the presignedURL field on the file query
// if the field is requested
func InterceptorPresignedURL() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.FileFunc(func(ctx context.Context, q *generated.FileQuery) (generated.Value, error) {
			logx.FromContext(ctx).Debug().Msg("InterceptorPresignedURL")

			if proxy.ShouldBypassPresignInterceptor(ctx) {
				return next.Query(ctx, q)
			}

			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			if q.ObjectManager == nil {
				logx.FromContext(ctx).Warn().Msg("object manager is nil, skipping presignedURL")

				return v, nil
			}

			// get the fields that were queried and check for the presignedURL field
			fields := graphutils.CheckForRequestedField(ctx, "presignedURL")

			// if the presignedURL field wasn't queried, return the result as is
			if !fields {
				return v, nil
			}

			// cast to the expected output format
			res, ok := v.([]*generated.File)
			if ok {
				for _, f := range res {
					if err := setPresignedURL(ctx, f, q); err != nil {
						logx.FromContext(ctx).Warn().Err(err).Msg("failed to set presignedURL")
					}
				}

				return v, nil
			}

			// if its not a list, check the single entry
			f, ok := v.(*generated.File)
			if ok {
				if err := setPresignedURL(ctx, f, q); err != nil {
					logx.FromContext(ctx).Warn().Err(err).Msg("failed to set presignedURLs")
				}

				return v, nil
			}

			return v, nil
		})
	})
}

// presignedURLDuration is the duration for the presigned URL to be valid
const presignedURLDuration = 60 * time.Minute * 24 // 24 hours

// setPresignedURL sets the presigned URL for the file response that is valid for 24 hours
func setPresignedURL(ctx context.Context, file *generated.File, q *generated.FileQuery) error {
	// if the storage path or file is empty, skip
	if file == nil || file.StoragePath == "" {
		return nil
	}

	// Convert ent File to storagetypes.File
	storageFile := &storagetypes.File{
		ID:           file.ID,
		OriginalName: file.ProvidedFileName,
		FileMetadata: storagetypes.FileMetadata{
			Key:          file.StoragePath,
			Bucket:       file.StorageVolume,
			Region:       file.StorageRegion,
			ContentType:  file.DetectedContentType,
			Size:         file.PersistedFileSize,
			ProviderType: storagetypes.ProviderType(file.StorageProvider),
			FullURI:      file.URI,
			ProviderHints: &storagetypes.ProviderHints{
				KnownProvider: storagetypes.ProviderType(file.StorageProvider),
			},
		},
	}

	// Convert metadata from map[string]interface{} to map[string]string
	if file.Metadata != nil {
		metadata := make(map[string]string)
		for k, v := range file.Metadata {
			if str, ok := v.(string); ok {
				metadata[k] = str
			}
		}

		storageFile.Metadata = metadata
	}

	url, err := q.ObjectManager.GetPresignedURL(ctx, storageFile, presignedURLDuration)
	if err != nil {
		logx.FromContext(ctx).Err(err).Msg("failed to get presigned URL")

		return err
	}

	file.PresignedURL = url

	return nil
}
