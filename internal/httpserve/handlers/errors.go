package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/internal/ent/generated"
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
func (h *Handler) InternalServerError(ctx echo.Context, err error) error {
	if err := ctx.JSON(http.StatusInternalServerError, rout.ErrorResponse(err)); err != nil {
		return err
	}

	return err
}

// Unauthorized returns a 401 Unauthorized response with the error message.
func (h *Handler) Unauthorized(ctx echo.Context, err error) error {
	if err := ctx.JSON(http.StatusUnauthorized, rout.ErrorResponse(err)); err != nil {
		return err
	}

	return err
}

// NotFound returns a 404 Not Found response with the error message.
func (h *Handler) NotFound(ctx echo.Context, err error) error {
	if err := ctx.JSON(http.StatusNotFound, rout.ErrorResponse(err)); err != nil {
		return err
	}

	return err
}

// BadRequest returns a 400 Bad Request response with the error message.
func (h *Handler) BadRequest(ctx echo.Context, err error) error {
	if err := ctx.JSON(http.StatusBadRequest, rout.ErrorResponse(err)); err != nil {
		return err
	}

	return err
}

// BadRequest returns a 400 Bad Request response with the error message.
func (h *Handler) BadRequestWithCode(ctx echo.Context, err error, code rout.ErrorCode) error {
	if err := ctx.JSON(http.StatusBadRequest, rout.ErrorResponseWithCode(err, code)); err != nil {
		return err
	}

	return err
}

// InvalidInput returns a 400 Bad Request response with the error message.
func (h *Handler) InvalidInput(ctx echo.Context, err error) error {
	if err := ctx.JSON(http.StatusBadRequest, rout.ErrorResponseWithCode(err, InvalidInputErrCode)); err != nil {
		return err
	}

	return err
}

// Conflict returns a 409 Conflict response with the error message.
func (h *Handler) Conflict(ctx echo.Context, err string, code rout.ErrorCode) error {
	if err := ctx.JSON(http.StatusConflict, rout.ErrorResponseWithCode(err, code)); err != nil {
		return err
	}

	return fmt.Errorf("%w: %v", ErrConflict, err)
}

// TooManyRequests returns a 429 Too Many Requests response with the error message.
func (h *Handler) TooManyRequests(ctx echo.Context, err error) error {
	if err := ctx.JSON(http.StatusTooManyRequests, rout.ErrorResponse(err)); err != nil {
		return err
	}

	return err
}

// Success returns a 200 OK response with the response object.
func (h *Handler) Success(ctx echo.Context, rep interface{}) error {
	return ctx.JSON(http.StatusOK, rep)
}

// Created returns a 201 Created response with the response object.
func (h *Handler) Created(ctx echo.Context, rep interface{}) error {
	return ctx.JSON(http.StatusCreated, rep)
}

func (h *Handler) SuccessBlob(ctx echo.Context, rep interface{}) error {
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
	ctx.Response().WriteHeader(http.StatusOK)

	encoder := json.NewEncoder(ctx.Response())
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("  ", "  ")

	return encoder.Encode(rep)
}

// Redirect returns a 302 Found response with the location header.
func (h *Handler) Redirect(ctx echo.Context, location string) error {
	return ctx.Redirect(http.StatusFound, location)
}
