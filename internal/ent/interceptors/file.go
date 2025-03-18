package interceptors

import (
	"context"
	"time"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"

	"github.com/theopenlane/gqlgen-plugins/graphutils"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
)

// InterceptorPresignedURL is an ent interceptor that sets the presignedURL field on the file query
// if the field is requested
func InterceptorPresignedURL() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return intercept.FileFunc(func(ctx context.Context, q *generated.FileQuery) (generated.Value, error) {
			v, err := next.Query(ctx, q)
			if err != nil {
				return nil, err
			}

			if q.ObjectManager == nil {
				log.Warn().Msg("object manager is nil, skipping presignedURL")

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
						log.Warn().Err(err).Msg("failed to set presignedURL")
					}
				}

				return v, nil
			}

			// if its not a list, check the single entry
			f, ok := v.(*generated.File)
			if ok {
				if err := setPresignedURL(ctx, f, q); err != nil {
					log.Warn().Err(err).Msg("failed to set presignedURLs")
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

	url, err := q.ObjectManager.Storage.GetPresignedURL(ctx, file.StoragePath, presignedURLDuration)
	if err != nil {
		log.Err(err).Msg("failed to get presigned URL")

		return err
	}

	file.PresignedURL = url

	return nil
}
