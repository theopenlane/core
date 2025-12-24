package common //nolint:revive

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gocarina/gocsv"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"github.com/vektah/gqlparser/v2/ast"

	"github.com/theopenlane/echox/middleware/echocontext"
	"github.com/theopenlane/gqlgen-plugins/graphutils"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/graphapi/gqlerrors"
	"github.com/theopenlane/core/internal/objects"
	"github.com/theopenlane/core/internal/objects/store"
	"github.com/theopenlane/core/internal/objects/upload"
	"github.com/theopenlane/core/pkg/logx"
	pkgobjects "github.com/theopenlane/core/pkg/objects"
)

const (
	// DefaultMaxMemoryMB is the default max memory for multipart forms (32MB)
	DefaultMaxMemoryMB = 32
	// GraphPath is the api graph path for the graphql server
	GraphPath = "query"
	// Default Complexity limit for graphql queries set by default, this can be overridden in the config
	DefaultComplexityLimit = 100
	// IntrospectionComplexity is the complexity limit for introspection queries
	IntrospectionComplexity = 200
)

// injectFileUploader adds the file uploader as middleware to the graphql operation
// this is used to handle file uploads to a storage backend, add the file to the file schema
// and add the uploaded files to the echo context
func injectFileUploader(u *objects.Service) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (any, error) {
		rctx := graphql.GetFieldContext(ctx)

		// if the field context is nil or its not a resolver, return the next handler
		if rctx == nil || !rctx.IsResolver {
			return next(ctx)
		}

		// if the field context is a resolver, handle the file uploads
		op := graphql.GetOperationContext(ctx)

		// only handle mutations because the file uploads are only in mutations
		if op.Operation.Operation != "mutation" {
			return next(ctx)
		}

		// Use consolidated parsing logic for GraphQL variables
		inputKey := graphutils.GetInputFieldVariableName(ctx)

		// Parse files from GraphQL variables using the consolidated parser
		filesMap, err := pkgobjects.ParseFilesFromSource(op.Variables)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to parse files from graphql variables")

			return nil, err
		}

		// Convert to flat list, filtering out input key and adding object details
		uploads := []pkgobjects.File{}
		for k, files := range filesMap {
			// skip the input key
			if k == inputKey {
				continue
			}

			for _, fileUpload := range files {
				// Buffer the file in memory if small enough, otherwise leave as-is
				if fileUpload.RawFile != nil {
					buffered, err := pkgobjects.NewBufferedReaderFromReader(fileUpload.RawFile)
					if err == nil {
						fileUpload.RawFile = buffered
					}
				}

				// Add object details using existing logic
				enhanced, err := retrieveObjectDetails(rctx, k, &fileUpload)
				if err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to retrieve object details for upload")

					return nil, err
				}

				uploads = append(uploads, *enhanced)
			}
		}

		// return the next handler if there are no uploads
		if len(uploads) == 0 {
			return next(ctx)
		}

		// Clean up any temporary files created by multipart form parser
		ec, err := echocontext.EchoContextFromContext(ctx)
		if err == nil && ec.Request().MultipartForm != nil {
			multipartForm := ec.Request().MultipartForm

			defer func() {
				if removeErr := multipartForm.RemoveAll(); removeErr != nil {
					logx.FromContext(ctx).Warn().Err(removeErr).Msg("failed to clean multipart form")
				}
			}()
		}

		ctx, _, err = upload.HandleUploads(ctx, u, uploads)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("failed to handle file uploads")

			return nil, err
		}

		// add the uploaded files to the echo context if there are any
		// this is useful for using other middleware that depends on the echo context
		// and the uploaded files (e.g. body dump middleware)
		if ec != nil {
			ec.SetRequest(ec.Request().WithContext(ctx))
		}

		// process the rest of the resolver
		field, err := next(ctx)
		if err != nil {
			// rollback the uploaded files in case of an error
			upload.HandleRollback(ctx, u, uploads)
			logx.FromContext(ctx).Error().Err(err).Msg("failed to process resolver after file upload, rolling back uploads")

			return nil, err
		}

		// add the file permissions before returning the field
		if ctx, err = store.AddFilePermissions(ctx); err != nil {
			// rollback the uploaded files in case of an error
			upload.HandleRollback(ctx, u, uploads)

			logx.FromContext(ctx).Error().Err(err).Msg("failed to add file permissions to uploaded files")

			return nil, err
		}

		return field, nil
	}
}

// UnmarshalBulkData unmarshals the input bulk data into a slice of the given type
func UnmarshalBulkData[T any](input graphql.Upload) ([]*T, error) {
	// read the csv file
	var data []*T

	stream, readErr := io.ReadAll(input.File)
	if readErr != nil {
		return nil, readErr
	}
	// Configure the CSV reader gocsv will use
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.LazyQuotes = true    // tolerate odd quotes
		r.FieldsPerRecord = -1 // allow variable field counts
		r.TrimLeadingSpace = true
		return r
	})

	if err := gocsv.UnmarshalBytes(stream, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// inputWithOwnerID is a struct that contains the owner id
// this is used to unmarshal the owner id from the input
type inputWithOwnerID struct {
	OwnerID *string `json:"ownerID"`
}

// GetOrgOwnerFromInput retrieves the owner id from the input
// input can be of any type, but must contain an owner id field
// if the owner id is not found, it returns nil
func GetOrgOwnerFromInput[T any](input *T) (*string, error) {
	if input == nil {
		return nil, nil
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	var ownerInput inputWithOwnerID
	if err := json.Unmarshal(inputBytes, &ownerInput); err != nil {
		return nil, err
	}

	return ownerInput.OwnerID, nil
}

// GetBulkUploadOwnerInput retrieves the owner id from the bulk upload input
// if there are multiple owner ids, it returns an error
// this is used to ensure that the owner id is consistent across all inputs
func GetBulkUploadOwnerInput[T any](input []*T) (*string, error) {
	var ownerID *string

	for _, i := range input {
		ownerInputID, err := GetOrgOwnerFromInput(i)
		if err != nil {
			return nil, err
		}

		if ownerInputID == nil {
			log.Error().Msg("owner id not found in bulk upload input")

			return nil, gqlerrors.NewCustomError(
				gqlerrors.BadRequestErrorCode,
				"unable to determine the organization owner id from the input, no owner id found",
				ErrNoOrganizationID,
			)
		}

		// if the owner doesn't match the previous owner, return an error
		if ownerID != nil && *ownerInputID != *ownerID {
			log.Error().Msg("multiple owner ids found in bulk upload input")

			return nil, gqlerrors.NewCustomError(
				gqlerrors.BadRequestErrorCode,
				"unable to determine the organization owner id from the input, multiple owner ids found",
				ErrNoOrganizationID,
			)
		}

		ownerID = ownerInputID
	}

	return ownerID, nil
}

// SetOrganizationInAuthContext sets the organization in the auth context based on the input if it is not already set
// in most cases this is a no-op because the organization id is set in the auth middleware
// only when multiple organizations are authorized (e.g. with a PAT) is this necessary
func SetOrganizationInAuthContext(ctx context.Context, inputOrgID *string) error {
	// if org is in context or the user is a system admin, return
	if ok, err := checkOrgInContext(ctx); ok && err == nil {
		return nil
	}

	return setOrgFromInputInContext(ctx, inputOrgID)
}

// SetOrganizationInAuthContextBulkRequest sets the organization in the auth context based on the input if it is not already set
// in most cases this is a no-op because the organization id is set in the auth middleware
// in the case of personal access tokens, this is necessary to ensure the organization id is set
// the organization must be the same across all inputs in the bulk request
func SetOrganizationInAuthContextBulkRequest[T any](ctx context.Context, input []*T) error {
	// if org is in context or the user is a system admin, return
	if ok, err := checkOrgInContext(ctx); ok && err == nil {
		return nil
	}

	ownerID, err := GetBulkUploadOwnerInput(input)
	if err != nil {
		return err
	}

	return setOrgFromInputInContext(ctx, ownerID)
}

// checkOrgInContext checks if the organization is already set in the context
// if the organization is set, it returns true
// if the user is a system admin, it also returns true
func checkOrgInContext(ctx context.Context) (bool, error) {
	// allow system admins to bypass the organization check
	isAdmin, err := rule.CheckIsSystemAdminWithContext(ctx)
	if err == nil && isAdmin {
		log.Debug().Bool("isAdmin", isAdmin).Msg("user is system admin, bypassing setting organization in auth context")

		return true, nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err == nil && orgID != "" {
		return true, nil
	}

	return false, nil
}

// setOrgFromInputInContext sets the organization in the auth context based on the input org ID, ensuring
// the org is authenticated and exists in the context
func setOrgFromInputInContext(ctx context.Context, inputOrgID *string) error {
	if inputOrgID == nil {
		// this would happen on a PAT authenticated request because the org id is not set
		return ErrNoOrganizationID
	}

	// ensure this org is authenticated
	orgIDs, err := auth.GetOrganizationIDsFromContext(ctx)
	if err != nil {
		return err
	}

	if !lo.Contains(orgIDs, *inputOrgID) {
		return fmt.Errorf("%w: organization id %s not found in the authenticated organizations", rout.ErrBadRequest, *inputOrgID)
	}

	err = auth.SetOrganizationIDInAuthContext(ctx, *inputOrgID)
	if err != nil {
		return err
	}

	return nil
}

// CheckAllowedAuthType checks how the user is authenticated and returns an error
// if the user is authenticated with an API token for a user owned setting
func CheckAllowedAuthType(ctx context.Context) error {
	ac, err := auth.GetAuthenticatedUserFromContext(ctx)
	if err != nil {
		return err
	}

	if ac.AuthenticationType == auth.APITokenAuthentication {
		return fmt.Errorf("%w: unable to use API token to update user settings", rout.ErrBadRequest)
	}

	return nil
}

// retrieveObjectDetails retrieves the object details from the field context
func retrieveObjectDetails(rctx *graphql.FieldContext, key string, upload *pkgobjects.File) (*pkgobjects.File, error) {
	// loop through the arguments in the request
	for _, arg := range rctx.Field.Arguments {
		// check if the argument is an upload
		if argIsUpload(arg) {
			// check if the argument name matches the key
			if arg.Name == key {
				objectID, ok := rctx.Args["id"].(string)
				if ok {
					upload.CorrelatedObjectID = objectID
					upload.Parent.ID = objectID
				}

				objectType := stripOperation(rctx.Field.Name)
				upload.CorrelatedObjectType = objectType
				// Parent.Type must use snake_case to align with authz tuple expectations (e.g., trust_center_doc)
				upload.Parent.Type = lo.SnakeCase(objectType)
				upload.FieldName = arg.Name
				upload.Key = arg.Name // Also set Key in FileMetadata for backwards compatibility

				return upload, nil
			}
		}
	}

	log.Debug().Str("key", key).Msg("unable to determine object type - no matching upload argument found")

	return upload, ErrUnableToDetermineObjectType
}

// argIsUpload checks if the argument is an upload
func argIsUpload(arg *ast.Argument) bool {
	if arg == nil || arg.Value == nil || arg.Value.ExpectedType == nil {
		return false
	}

	if arg.Value.ExpectedType.NamedType == "Upload" {
		return true
	}

	if arg.Value.ExpectedType.Elem != nil && arg.Value.ExpectedType.Elem.NamedType == "Upload" {
		return true
	}

	return false
}

// stripOperation strips the operation prefix from the field name and returns the remainder unchanged
// e.g. updateUser becomes User, createTrustCenterDoc becomes TrustCenterDoc
func stripOperation(field string) string {
	operations := []string{"create", "update", "delete", "get"}

	result := field
	for _, op := range operations {
		if strings.HasPrefix(result, op) {
			result = strings.ReplaceAll(result, op, "")

			break
		}
	}

	return strings.TrimPrefix(result, "Upload")
}

// IsEmpty checks if the given interface is empty
func IsEmpty(x any) bool {
	if x == nil {
		return true
	}

	switch v := x.(type) {
	case string:
		return v == lo.Empty[string]()
	case []int, []string, []float64, []any:
		return isEmptySlice(v)
	case map[string]any:
		return len(v) == 0
	case int, int8, int16, int32, int64:
		return v == lo.Empty[int]()
	case uint, uint8, uint16, uint32, uint64:
		return v == lo.Empty[uint]()
	case float32, float64:
		return v == lo.Empty[float64]()
	case bool:
		return v == lo.Empty[bool]()
	default:
		// fallback to this helper, which expects a comparable type
		return lo.IsNil(v)
	}

}

// isEmptySlice checks if the given interface is an empty slice
func isEmptySlice(x any) bool {
	switch v := x.(type) {
	case []int:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case []float64:
		return len(v) == 0
	case []any:
		return len(v) == 0
	default:
		return false
	}
}

// ConvertToObject converts an object to a specific type
func ConvertToObject[J any](obj any) (*J, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var result J
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
