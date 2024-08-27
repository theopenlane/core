package tokens

import (
	"errors"
	"fmt"
)

// Error constants
var (
	// ErrTokenManagerFailedInit returns when the token manager was not correctly provided signing keys
	ErrTokenManagerFailedInit = errors.New("token manager not initialized with signing keys")

	// ErrFailedRetrieveClaimsFromToken returns when claims can not be retrieved from an access token
	ErrFailedRetrieveClaimsFromToken = errors.New("could not retrieve claims from access token")

	// ErrTokenMissingKid returns when the kid cannot be found in the header of the token
	ErrTokenMissingKid = errors.New("token does not have kid in header")

	// ErrFailedParsingKid returns when the kid could not be parsed
	ErrFailedParsingKid = errors.New("could not parse kid: %s")

	// ErrUnknownSigningKey returns when the signing key fetched does not match the loaded managed keys
	ErrUnknownSigningKey = errors.New("unknown signing key")
)

var (
	// The below block of error constants are only used in comparison(s) checks with ValidationError and are the inner standard errors which could occur when performing token validation; these won't be referenced in return error messages within the code directly and wrapped with our custom errors

	// ErrTokenMalformed returns when a token is malformed
	ErrTokenMalformed = errors.New("token is malformed")

	// ErrTokenUnverifiable is returned when the token could not be verified because of signing problems
	ErrTokenUnverifiable = errors.New("token is unverifiable")

	// ErrTokenSignatureInvalid  is returned when the signature is invalid
	ErrTokenSignatureInvalid = errors.New("token signature is invalid")

	// ErrTokenInvalidAudience is returned when AUD validation failed
	ErrTokenInvalidAudience = errors.New("token has invalid audience")

	// ErrTokenExpired  is returned when EXP validation failed
	ErrTokenExpired = errors.New("token is expired")

	// ErrTokenUsedBeforeIssued is returned when the token is used before issued
	ErrTokenUsedBeforeIssued = errors.New("token used before issued")

	// ErrTokenInvalidIssuer is returned when ISS validation failed
	ErrTokenInvalidIssuer = errors.New("token has invalid issuer")

	// ErrTokenNotValidYet is returned when NBF validation failed
	ErrTokenNotValidYet = errors.New("token is not valid yet")

	// ErrTokenNotValid is returned when the token is invalid
	ErrTokenNotValid = errors.New("token is invalid")

	// ErrTokenInvalidID is returned when the token has an invalid id
	ErrTokenInvalidID = errors.New("token has invalid id")

	// ErrTokenInvalidClaims is returned when the token has invalid claims
	ErrTokenInvalidClaims = errors.New("token has invalid claims")

	// ErrMissingEmail is returned when the token is attempted to be verified but the email is missing
	ErrMissingEmail = errors.New("unable to create verification token, email is missing")

	// ErrTokenMissingEmail is returned when the verification is missing an email address
	ErrTokenMissingEmail = errors.New("email verification token is missing email address")

	// ErrInvalidSecret is returned when the verification contains of secret of invalid length
	ErrInvalidSecret = errors.New("email verification token contains an invalid secret")

	// ErrMissingUserID is returned when a reset token is trying to be created but no user id is provided
	ErrMissingUserID = errors.New("unable to create reset token, user id is required")

	// ErrTokenMissingUserID is returned when the reset token is missing the required user id
	ErrTokenMissingUserID = errors.New("reset token is missing user id")

	// ErrInviteTokenMissingOrgID is returned when the invite token is missing the org owner ID match
	ErrInviteTokenMissingOrgID = errors.New("invite token is missing org id")

	// ErrInviteTokenMissingEmail
	ErrInviteTokenMissingEmail = errors.New("invite token is missing email")

	// ErrExpirationIsRequired is returned when signing info is provided a zero-value expiration
	ErrExpirationIsRequired = errors.New("signing info requires a non-zero expiration")

	// ErrFailedSigning is returned when an error occurs when trying to generate signing info with expiration
	ErrFailedSigning = errors.New("error occurred when attempting to signing info")

	// ErrTokenInvalid is returned when unable to verify the token with the signature and secret provided
	ErrTokenInvalid = errors.New("unable to verify token")
)

// The errors that might occur when parsing and validating a token
const (
	// ValidationErrorMalformed is returned when the token is malformed
	ValidationErrorMalformed uint32 = 1 << iota

	// ValidationErrorUnverifiableToken is returned when the token could not be verified because of signing problems
	ValidationErrorUnverifiable

	// ValidationErrorSignatureInvalid is returned when the signature is invalid
	ValidationErrorSignatureInvalid

	// ValidationErrorAudience is returned when AUD validation failed
	ValidationErrorAudience

	// ValidationErrorExpired is returned when EXP validation failed
	ValidationErrorExpired

	// ValidationErrorIssuedAt is returned when IAT validation failed
	ValidationErrorIssuedAt

	// ValidationErrorIssuer is returned when ISS validation failed
	ValidationErrorIssuer

	// ValidationErrorNotValidYet is returned when NBF validation failed
	ValidationErrorNotValidYet

	// ValidationErrorID is returned when JTI validation failed
	ValidationErrorID

	// ValidationErrorClaimsInvalid is returned when there is a generic claims validation failure
	ValidationErrorClaimsInvalid
)

// NewValidationError is a helper for constructing a ValidationError with a string error message
func NewValidationError(errorText string, errorFlags uint32) *ValidationError {
	return &ValidationError{
		text:   errorText,
		Errors: errorFlags,
	}
}

// ValidationError represents an error from Parse if token is not valid
type ValidationError struct {
	Inner  error
	Errors uint32
	text   string
}

// Error is the implementation of the err interface for ValidationError
func (e ValidationError) Error() string {
	i := e.Inner

	switch {
	case i != nil:
		return e.Inner.Error()
	case e.text != "":
		return e.text
	default:
		return "token is invalid"
	}
}

// Unwrap gives errors.Is and errors.As access to the inner errors defined above
func (e *ValidationError) Unwrap() error {
	return e.Inner
}

// Is checks if this ValidationError is of the supplied error. We are first checking for the exact
// error message by comparing the inner error message. If that fails, we compare using the error
// flags. This way we can use custom error messages and leverage errors.Is using the global error
// variables, plus I just learned how to use errors.Is today so this is pretty sweet
func (e *ValidationError) Is(err error) bool {
	// Check, if our inner error is a direct match
	if errors.Is(errors.Unwrap(e), err) {
		return true
	}

	// Otherwise, we need to match using our error flags
	switch err {
	case ErrTokenMalformed:
		return e.Errors&ValidationErrorMalformed != 0
	case ErrTokenUnverifiable:
		return e.Errors&ValidationErrorUnverifiable != 0
	case ErrTokenSignatureInvalid:
		return e.Errors&ValidationErrorSignatureInvalid != 0
	case ErrTokenInvalidAudience:
		return e.Errors&ValidationErrorAudience != 0
	case ErrTokenExpired:
		return e.Errors&ValidationErrorExpired != 0
	case ErrTokenUsedBeforeIssued:
		return e.Errors&ValidationErrorIssuedAt != 0
	case ErrTokenInvalidIssuer:
		return e.Errors&ValidationErrorIssuer != 0
	case ErrTokenNotValidYet:
		return e.Errors&ValidationErrorNotValidYet != 0
	case ErrTokenInvalidID:
		return e.Errors&ValidationErrorID != 0
	case ErrTokenInvalidClaims:
		return e.Errors&ValidationErrorClaimsInvalid != 0
	}

	return false
}

// ParseError is defining a custom error type called `ParseError`
type ParseError struct {
	Object string
	Value  string
	Err    error
}

// Error returns the ParseError in string format
func (e *ParseError) Error() string {
	return fmt.Sprintf("could not parse %s %s: %v", e.Object, e.Value, e.Err)
}

// The function newParseError creates a new ParseError object with the given parameters
func newParseError(o string, v string, err error) *ParseError {
	return &ParseError{
		Object: o,
		Value:  v,
		Err:    err,
	}
}
