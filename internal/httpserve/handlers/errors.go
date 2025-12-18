package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/metrics"
	models "github.com/theopenlane/core/pkg/openapi"
)

var (
	// ErrBadRequest is returned when the request cannot be processed
	ErrBadRequest = errors.New("invalid request")
	// ErrProcessingRequest is returned when the request cannot be processed
	ErrProcessingRequest = errors.New("error processing request, please try again")
	// ErrMissingRequiredFields is returned when the login request has an empty username or password
	ErrMissingRequiredFields = errors.New("invalid request, missing username and/or password")
	// ErrInvalidInput is returned when the input is invalid
	ErrInvalidInput = errors.New("invalid input")
	// ErrNotFound is returned when the requested object is not found
	ErrNotFound = errors.New("object not found in the database")
	// ErrMissingField is returned when a field is missing duh
	ErrMissingField = errors.New("missing required field")
	// ErrInvalidCredentials is returned when the password is invalid or missing
	ErrInvalidCredentials = errors.New("credentials are missing or invalid")
	// ErrUnverifiedUser is returned when email_verified on the user is false
	ErrUnverifiedUser = errors.New("user is not verified")
	// ErrUnableToVerifyEmail is returned when user's email is not able to be verified
	ErrUnableToVerifyEmail = errors.New("could not verify email")
	// ErrMaxAttempts is returned when user has requested the max retry attempts to verify their email
	ErrMaxAttempts = errors.New("max attempts verifying email address")
	// ErrNoEmailFound is returned when using an oauth provider and the email address cannot be determined
	ErrNoEmailFound = errors.New("no email found from oauth provider")
	// ErrInvalidProvider is returned when registering a user with an unsupported oauth provider
	ErrInvalidProvider = errors.New("oauth2 provider not supported")
	// ErrProviderDisabled is returned when the provider is configured but inactive
	ErrProviderDisabled = errors.New("provider is currently disabled")
	// ErrMissingOrganizationContext is returned when organization context is missing during OAuth flow
	ErrMissingOrganizationContext = errors.New("missing organization context")
	// ErrInvalidOrganizationContext is returned when organization context does not match authenticated user
	ErrInvalidOrganizationContext = errors.New("invalid organization context")
	// ErrMissingUserContext is returned when user context is missing during OAuth flow
	ErrMissingUserContext = errors.New("missing user context")
	// ErrInvalidUserContext is returned when user context does not match authenticated user
	ErrInvalidUserContext = errors.New("invalid user context")
	// ErrNoAuthUser is returned when the user couldn't be identified by the request
	ErrNoAuthUser = errors.New("could not identify authenticated user in request")
	// ErrPassWordResetTokenInvalid is returned when the provided token and secret do not match the stored
	ErrPassWordResetTokenInvalid = errors.New("password reset token invalid")
	// ErrNonUniquePassword is returned when the password was already used
	ErrNonUniquePassword = errors.New("password was already used, please try again")
	// ErrPasswordTooWeak is returned when the password is too weak
	ErrPasswordTooWeak = errors.New("password is too weak: use a combination of upper and lower case letters, numbers, and special characters")
	// ErrMaxDeviceLimit is returned when the user has reached the max device limit
	ErrMaxDeviceLimit = errors.New("max device limit reached")
	// ErrDeviceAlreadyRegistered is returned when the device is already registered
	ErrDeviceAlreadyRegistered = errors.New("device already registered")
	// ErrSubscriberNotFound is returned when the subscriber is not found
	ErrSubscriberNotFound = errors.New("subscriber not found")
	// ErrExpiredToken is returned when the token has expired
	ErrExpiredToken = errors.New("token has expired")
	// ErrUnauthorized is returned when the user is not authorized to make the request
	ErrUnauthorized = errors.New("not authorized")
	// ErrConflict is returned when the request cannot be processed due to a conflict
	ErrConflict = errors.New("conflict")
	// ErrAlreadySwitchedIntoOrg is returned when a user attempts to switch into an org they are currently authenticated in
	ErrAlreadySwitchedIntoOrg = errors.New("user already switched into organization")
	// ErrNoBillingEmail is returned when the user has no billing email
	ErrNoBillingEmail = errors.New("no billing email found")
	// ErrPersonalOrgsNoBilling is returned when the org ID looked up is a personal org
	ErrPersonalOrgsNoBilling = errors.New("personal orgs do not have billing")
	// ErrInvalidRecoveryCode is returned when the recovery code is invalid
	ErrInvalidRecoveryCode = errors.New("invalid code provided")
	// ErrUnsupportedEventType is returned when the event type is not supported
	ErrUnsupportedEventType = errors.New("unsupported event type")
	// ErrUnableToRegisterJobRunner is returned when the job runner node cannot be registered
	ErrUnableToRegisterJobRunner = errors.New("could not register your job runner at this time")
	// ErrJobRunnerRegistrationTokenExpired is returned when a token has expired
	ErrJobRunnerRegistrationTokenExpired = errors.New("job runner registration token expired")
	// ErrJobRunnerAlreadyRegistered is returned when we hit the ip address unique constraint
	ErrJobRunnerAlreadyRegistered = errors.New("this job runner node exists and cannot be registered twice")
	// ErrMissingOIDCConfig is returned when the OIDC configuration is missing
	ErrMissingOIDCConfig = errors.New("missing OIDC configuration, please contact support")
	// ErrStateMismatch is returned when the state parameter does not match the expected value
	ErrStateMismatch = errors.New("state parameter does not match, possible CSRF attack or session expired")
	// ErrMissingSSOConfig is returned when the SSO configuration is missing
	ErrNonceMissing = errors.New("missing nonce cookie, possible CSRF attack or session expired")
	// ErrMissingReferer is returned when the referer is missing from the request
	ErrMissingReferer = errors.New("referer is required")
	// ErrInvalidRefererURL is returned when the referer URL is invalid
	ErrInvalidRefererURL = errors.New("invalid referer URL")
	// ErrMissingSlugInPath is returned when the slug is missing from the path
	ErrMissingSlugInPath = errors.New("slug is required in the path for default trust center domain")
	// ErrTrustCenterNotFound is returned when the trust center is not found
	ErrTrustCenterNotFound = errors.New("trust center not found")
	// ErrAssessmentNotFound is returned when the assessment is not found
	ErrAssessmentNotFound = errors.New("assessment not found")
	// ErrTemplateNotFound is returned when the template is not found
	ErrTemplateNotFound = errors.New("template not found")
	// ErrMissingQuestionnaireContext is returned when the questionnaire context is missing
	ErrMissingQuestionnaireContext = errors.New("missing questionnaire context")
	// ErrMissingAssessmentID is returned when the assessment ID is missing from the context
	ErrMissingAssessmentID = errors.New("missing assessment ID")
	// ErrMissingQuestionnaireData is returned when the questionnaire data is missing from the request
	ErrMissingQuestionnaireData = errors.New("missing questionnaire data")
	// ErrMissingEmail is returned when the email is missing from the context
	ErrMissingEmail = errors.New("missing email")
	// ErrAssessmentResponseNotFound is returned when the assessment response is not found
	ErrAssessmentResponseNotFound = errors.New("assessment response not found")
	// ErrAssessmentResponseAlreadyCompleted is returned when trying to submit a questionnaire that has already been completed
	ErrAssessmentResponseAlreadyCompleted = errors.New("assessment response has already been completed")
	// ErrAssessmentResponseOverdue is returned when trying to access or submit a questionnaire that is past due
	ErrAssessmentResponseOverdue = errors.New("assessment response is overdue")
	// ErrAuthenticationRequired indicates that the user must be authenticated to perform this action
	ErrAuthenticationRequired = errors.New("authentication required")
	// ErrNoActiveImpersonationSession indicates that there is no active impersonation session
	ErrNoActiveImpersonationSession = errors.New("no active impersonation session")
	// ErrInvalidSessionID indicates that the provided session ID is invalid
	ErrInvalidSessionID = errors.New("invalid session ID")
	// ErrInsufficientPermissionsSupport indicates that the user does not have permissions to perform support impersonation
	ErrInsufficientPermissionsSupport = errors.New("insufficient permissions for support impersonation")
	// ErrInsufficientPermissionsAdmin indicates that the user does not have permissions to perform admin impersonation
	ErrInsufficientPermissionsAdmin = errors.New("insufficient permissions for admin impersonation")
	// ErrJobImpersonationAdminOnly indicates that job impersonation is only allowed for system admins
	ErrJobImpersonationAdminOnly = errors.New("job impersonation only allowed for system admins")
	// ErrInvalidImpersonationType indicates that the provided impersonation type is invalid
	ErrInvalidImpersonationType = errors.New("invalid impersonation type")
	// ErrTargetUserNotFound indicates that the target user for impersonation was not found
	ErrTargetUserNotFound = errors.New("target user not found")
	// ErrTokenManagerNotConfigured indicates that the token manager is not configured
	ErrTokenManagerNotConfigured = errors.New("token manager not configured")
	// ErrFailedToExtractSessionID indicates that the session ID could not be extracted from the token
	ErrFailedToExtractSessionID = errors.New("failed to extract session ID from token")
	// ErrObjectStoreUnavailable indicates that the object store is not configured
	ErrObjectStoreUnavailable = errors.New("object store unavailable")
	// ErrDownloadTokenMissingFile indicates that the download token does not have an associated file
	ErrDownloadTokenMissingFile = errors.New("download token missing associated file")
	// ErrLoginFailed is returned when login fails
	ErrLoginFailed = errors.New("login failed, please check your credentials and try again")
	// ErrUnableToVerifyToken is returned when unable to verify a token
	ErrUnableToVerifyToken = errors.New("unable to verify token, please try again")
)

var (
	// DeviceRegisteredErrCode is returned when the device is already registered
	DeviceRegisteredErrCode rout.ErrorCode = "DEVICE_REGISTERED"
	// UserExistsErrCode is returned when the user already exists
	UserExistsErrCode rout.ErrorCode = "USER_EXISTS"
	// InvalidInputErrCode is returned when the input is invalid
	InvalidInputErrCode rout.ErrorCode = "INVALID_INPUT"
)

// IsConstraintError returns true if the error resulted from a database constraint violation.
func IsConstraintError(err error) bool {
	var e *generated.ConstraintError
	return errors.As(err, &e) || IsUniqueConstraintError(err) || IsForeignKeyConstraintError(err)
}

// IsUniqueConstraintError reports if the error resulted from a DB uniqueness constraint violation.
// e.g. duplicate value in unique index.
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	for _, s := range []string{
		"Error 1062",                 // MySQL
		"violates unique constraint", // Postgres
		"UNIQUE constraint failed",   // SQLite
	} {
		if strings.Contains(err.Error(), s) {
			return true
		}
	}

	return false
}

// IsForeignKeyConstraintError reports if the error resulted from a database foreign-key constraint violation.
// e.g. parent row does not exist.
func IsForeignKeyConstraintError(err error) bool {
	if err == nil {
		return false
	}

	for _, s := range []string{
		"Error 1451",                      // MySQL (Cannot delete or update a parent row).
		"Error 1452",                      // MySQL (Cannot add or update a child row).
		"violates foreign key constraint", // Postgres
		"FOREIGN KEY constraint failed",   // SQLite
	} {
		if strings.Contains(err.Error(), s) {
			return true
		}
	}

	return false
}

func invalidInputError(err error) error {
	field := err.(*generated.ValidationError).Name

	return fmt.Errorf("%w: %s was invalid", ErrInvalidInput, field)
}

// InternalServerError returns a 500 Internal Server Error response with the error message.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) InternalServerError(ctx echo.Context, err error, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerError(http.StatusInternalServerError)

	// Create error response
	errorResponse := rout.ErrorResponse(err)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		if schemaRef, schemaErr := openapi[0].Registry.GetOrRegister(errorResponse); schemaErr == nil {
			response := openapi3.NewResponse().
				WithDescription("Internal Server Error").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusInternalServerError, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["error"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(errorResponse))}
		}
	}

	return ctx.JSON(http.StatusInternalServerError, errorResponse)
}

// Unauthorized returns a 401 Unauthorized response with the error message.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) Unauthorized(ctx echo.Context, err error, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerError(http.StatusUnauthorized)

	// Create error response
	errorResponse := rout.ErrorResponse(err)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		if schemaRef, schemaErr := openapi[0].Registry.GetOrRegister(errorResponse); schemaErr == nil {
			response := openapi3.NewResponse().
				WithDescription("Unauthorized").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusUnauthorized, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["error"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(errorResponse))}
		}
	}

	return ctx.JSON(http.StatusUnauthorized, errorResponse)
}

// NotFound returns a 404 Not Found response with the error message.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) NotFound(ctx echo.Context, err error, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerError(http.StatusNotFound)

	// Create error response
	errorResponse := rout.ErrorResponse(err)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		if schemaRef, schemaErr := openapi[0].Registry.GetOrRegister(errorResponse); schemaErr == nil {
			response := openapi3.NewResponse().
				WithDescription("Not Found").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusNotFound, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["error"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(errorResponse))}
		}
	}

	return ctx.JSON(http.StatusNotFound, errorResponse)
}

// BadRequest returns a 400 Bad Request response with the error message.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) BadRequest(ctx echo.Context, err error, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerError(http.StatusBadRequest)

	// Create error response
	errorResponse := rout.ErrorResponse(err)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		if schemaRef, schemaErr := openapi[0].Registry.GetOrRegister(errorResponse); schemaErr == nil {
			response := openapi3.NewResponse().
				WithDescription("Bad Request").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusBadRequest, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["error"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(errorResponse))}
		}
	}

	return ctx.JSON(http.StatusBadRequest, errorResponse)
}

// BadRequestWithCode returns a 400 Bad Request response with the error message and code.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) BadRequestWithCode(ctx echo.Context, err error, code rout.ErrorCode, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerError(http.StatusBadRequest)

	// Create error response
	errorResponse := rout.ErrorResponseWithCode(err, code)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		if schemaRef, schemaErr := openapi[0].Registry.GetOrRegister(errorResponse); schemaErr == nil {
			response := openapi3.NewResponse().
				WithDescription("Bad Request").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusBadRequest, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["error"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(errorResponse))}
		}
	}

	return ctx.JSON(http.StatusBadRequest, errorResponse)
}

// InvalidInput returns a 400 Bad Request response with the error message.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) InvalidInput(ctx echo.Context, err error, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerError(http.StatusBadRequest)

	// Create error response
	errorResponse := rout.ErrorResponseWithCode(err, InvalidInputErrCode)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		if schemaRef, schemaErr := openapi[0].Registry.GetOrRegister(errorResponse); schemaErr == nil {
			response := openapi3.NewResponse().
				WithDescription("Invalid Input").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusBadRequest, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["error"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(errorResponse))}
		}
	}

	return ctx.JSON(http.StatusBadRequest, errorResponse)
}

// Conflict returns a 409 Conflict response with the error message.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) Conflict(ctx echo.Context, err string, code rout.ErrorCode, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerError(http.StatusConflict)

	// Create error response
	errorResponse := rout.ErrorResponseWithCode(err, code)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		if schemaRef, schemaErr := openapi[0].Registry.GetOrRegister(errorResponse); schemaErr == nil {
			response := openapi3.NewResponse().
				WithDescription("Conflict").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusConflict, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["error"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(errorResponse))}
		}
	}

	if jsonErr := ctx.JSON(http.StatusConflict, errorResponse); jsonErr != nil {
		return jsonErr
	}

	return fmt.Errorf("%w: %v", ErrConflict, err)
}

// TooManyRequests returns a 429 Too Many Requests response with the error message.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) TooManyRequests(ctx echo.Context, err error, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerError(http.StatusTooManyRequests)

	// Create error response
	errorResponse := rout.ErrorResponse(err)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		if schemaRef, schemaErr := openapi[0].Registry.GetOrRegister(errorResponse); schemaErr == nil {
			response := openapi3.NewResponse().
				WithDescription("Too Many Requests").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusTooManyRequests, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["error"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(errorResponse))}
		}
	}

	return ctx.JSON(http.StatusTooManyRequests, errorResponse)
}

// Success returns a 200 OK response with the response object.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) Success(ctx echo.Context, rep any, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerResult("Success", true)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		var exampleObject = rep

		// If the response implements ExampleProvider, use its example for the OpenAPI spec
		if provider, ok := any(&rep).(models.ExampleProvider); ok {
			exampleObject = provider.ExampleResponse()
		}

		if schemaRef, err := openapi[0].Registry.GetOrRegister(rep); err == nil {
			response := openapi3.NewResponse().
				WithDescription("Success").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusOK, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(exampleObject))}
		}
	}

	return ctx.JSON(http.StatusOK, rep)
}

// Created returns a 201 Created response with the response object.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) Created(ctx echo.Context, rep any, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerResult("Created", true)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		var exampleObject = rep

		// If the response implements ExampleProvider, use its example for the OpenAPI spec
		if provider, ok := any(&rep).(models.ExampleProvider); ok {
			exampleObject = provider.ExampleResponse()
		}

		if schemaRef, err := openapi[0].Registry.GetOrRegister(rep); err == nil {
			response := openapi3.NewResponse().
				WithDescription("Created").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusCreated, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(exampleObject))}
		}
	}

	return ctx.JSON(http.StatusCreated, rep)
}

// SuccessBlob returns a 200 OK response with the response object as pretty-printed JSON.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) SuccessBlob(ctx echo.Context, rep any, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerResult("SuccessBlob", true)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		if schemaRef, err := openapi[0].Registry.GetOrRegister(rep); err == nil {
			response := openapi3.NewResponse().
				WithDescription("Success").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusOK, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["success"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(rep))}
		}
	}

	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	ctx.Response().WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(ctx.Response())
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("  ", "  ")

	return encoder.Encode(rep)
}

// Forbidden returns a 403 Forbidden response with the error message.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) Forbidden(ctx echo.Context, err error, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerError(http.StatusForbidden)

	// Create error response
	errorResponse := rout.ErrorResponse(err)

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		if schemaRef, schemaErr := openapi[0].Registry.GetOrRegister(errorResponse); schemaErr == nil {
			response := openapi3.NewResponse().
				WithDescription("Forbidden").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusForbidden, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["error"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(errorResponse))}
		}
	}

	return ctx.JSON(http.StatusForbidden, errorResponse)
}

// Redirect returns a 302 Found response with the location header.
// Automatically registers the response schema if an OpenAPI context is provided.
func (h *Handler) Redirect(ctx echo.Context, location string, openapi ...*OpenAPIContext) error {
	// Record metrics
	metrics.RecordHandlerResult("Redirect", true)

	// Create redirect response object for schema registration
	redirectResponse := struct {
		Location string `json:"location" description:"The URL to redirect to"`
		Message  string `json:"message" description:"Redirect message"`
	}{
		Location: location,
		Message:  "Redirecting",
	}

	// Automatically register response schema if OpenAPI context is provided
	if len(openapi) > 0 && openapi[0] != nil && openapi[0].Operation != nil && openapi[0].Registry != nil {
		if schemaRef, schemaErr := openapi[0].Registry.GetOrRegister(redirectResponse); schemaErr == nil {
			response := openapi3.NewResponse().
				WithDescription("Redirect").
				WithContent(openapi3.NewContentWithJSONSchemaRef(schemaRef))
			openapi[0].Operation.AddResponse(http.StatusFound, response)

			response.Content.Get(httpsling.ContentTypeJSON).Examples = make(map[string]*openapi3.ExampleRef)
			response.Content.Get(httpsling.ContentTypeJSON).Examples["redirect"] = &openapi3.ExampleRef{Value: openapi3.NewExample(normalizeExampleValue(redirectResponse))}
		}
	}

	return ctx.Redirect(http.StatusFound, location)
}
