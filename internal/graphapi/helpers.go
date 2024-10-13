package graphapi

import (
	"context"
	"fmt"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gocarina/gocsv"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/middleware/objects"
	"github.com/theopenlane/core/pkg/events/soiree"
	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
	sliceutil "github.com/theopenlane/utils/slice"
)

const (
	// defaultMaxWorkers is the default number of workers in the pond pool when the pool was not created on server startup
	defaultMaxWorkers = 10
	// defaultMaxCapacity is the default capacity of the pond pool when the pool was not created on server startup
	defaultMaxCapacity = 100
)

// withTransactionalMutation automatically wrap the GraphQL mutations with a database transaction.
// This allows the ent.Client to commit at the end, or rollback the transaction in case of a GraphQL error.
func withTransactionalMutation(ctx context.Context) *ent.Client {
	return ent.FromContext(ctx)
}

// injectClient adds the db client to the context to be used with transactional mutations
func injectClient(db *ent.Client) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		ctx = ent.NewContext(ctx, db)
		return next(ctx)
	}
}

// injectFileUploader adds the file uploader as middleware to the graphql operation
// this is used to handle file uploads to a storage backend, add the file to the file schema
// and add the uploaded files to the echo context
func injectFileUploader(u *objects.Upload) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		op := graphql.GetOperationContext(ctx)

		// get the uploads from the variables
		// gqlgen will parse the variables and convert the graphql.Upload to a struct with the file data
		uploads := []objects.FileUpload{}
		for k, v := range op.Variables {
			up, ok := v.(graphql.Upload)
			if ok {
				fileUpload := objects.FileUpload{
					File:        up.File,
					Filename:    up.Filename,
					Size:        up.Size,
					ContentType: up.ContentType,
					Key:         k,
				}

				uploads = append(uploads, fileUpload)
			}
		}

		// return the next handler if there are no uploads
		if len(uploads) == 0 {
			return next(ctx)
		}

		// handle the file uploads
		ctx, err := u.FileUpload(ctx, uploads)
		if err != nil {
			return errorResponse(err)
		}

		// add the uploaded files to the echo context if there are any
		// this is useful for using other middleware that depends on the echo context
		// and the uploaded files (e.g. body dump middleware)
		ec, err := echocontext.EchoContextFromContext(ctx)
		if err == nil {
			ec.SetRequest(ec.Request().WithContext(ctx))
		}

		return next(ctx)
	}
}

// withPool returns the existing pool or creates a new one if it does not exist
func (r *queryResolver) withPool() *soiree.PondPool {
	if r.pool != nil {
		return r.pool
	}

	r.pool = soiree.NewPondPool(defaultMaxWorkers, defaultMaxCapacity)

	return r.pool
}

// unmarshalBulkData unmarshals the input bulk data into a slice of the given type
func unmarshalBulkData[T any](input graphql.Upload) ([]*T, error) {
	// read the csv file
	var data []*T
	stream, readErr := io.ReadAll(input.File)
	if readErr != nil {
		return nil, readErr
	}

	// parse the csv
	if err := gocsv.UnmarshalBytes(stream, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// setOrganizationInAuthContext sets the organization in the auth context based on the input if it is not already set
// in most cases this is a no-op because the organization id is set in the auth middleware
// only when multiple organizations are authorized (e.g. with a PAT) is this necessary
func setOrganizationInAuthContext(ctx context.Context, inputOrgID *string) error {
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err == nil && orgID != "" {
		return nil
	}

	if inputOrgID == nil {
		// this would happen on a PAT authenticated request because the org id is not set
		return fmt.Errorf("unable to determine organization id")
	}

	// ensure this org is authenticated
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return err
	}

	if !sliceutil.Contains(orgIDs, *inputOrgID) {
		return fmt.Errorf("organization id %s not found in the authenticated organizations", orgID)
	}

	au, err := auth.GetAuthenticatedUserContext(ctx)
	if err != nil {
		return err
	}

	au.OrganizationID = *inputOrgID

	ec, err := echocontext.EchoContextFromContext(ctx)
	if err != nil {
		return err
	}

	auth.SetAuthenticatedUserContext(ec, au)

	return nil
}

// checkAllowedAuthType checks how the user is authenticated and returns an error
// if the user is authenticated with an API token for a user owned setting
func checkAllowedAuthType(ctx context.Context) error {
	ac, err := auth.GetAuthenticatedUserContext(ctx)
	if err != nil {
		return err
	}

	if ac.AuthenticationType == auth.APITokenAuthentication {
		return fmt.Errorf("%w: unable to use API token to update user settings", rout.ErrBadRequest)
	}

	return nil
}
