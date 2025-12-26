package gqlerrors

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/theopenlane/common/models"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const (
	// ExtensionCodeKey is the key for the error code in the extensions
	ExtensionCodeKey = "code"
	// ExtensionMessageKey is the key for the error message in the extensions
	ExtensionMessageKey = "message"
	// ExtensionModuleKey is the key for the module that is required to get access to
	// the feature behind the graphql query
	ExtensionModuleKey = "module"
)

// CustomErrorType is an interface that defines a custom error type
type CustomErrorType interface {
	error
	// Code returns the static error code for the error in the gql extensions
	Code() string
	// Message returns the detailed error message for the error in the gql extensions
	Message() string
	// Module returns the module that failed checks for the query to succeed
	Module() models.OrgModule
}

var _ CustomErrorType = (*CustomError)(nil)

// CustomError is a struct that implements the CustomErrorType interface
type CustomError struct {
	code    string
	message string
	err     error
	module  models.OrgModule
}

// Error satisfies the CustomErrorType interface
func (e CustomError) Error() string {
	return e.err.Error()
}

// Code satisfies the CustomErrorType interface
func (e CustomError) Code() string {
	return e.code
}

// Message satisfies the CustomErrorType interface
func (e CustomError) Message() string {
	return e.message
}

// Module satisfies the CustomError interface
func (e CustomError) Module() models.OrgModule {
	return e.module
}

// NewCustomErrorWithModule creates a custom error with the given code, error and module
func NewCustomErrorWithModule(code, message string, err error, module *models.OrgModule) CustomError {
	cErr := CustomError{
		code:    code,
		message: message,
		err:     err,
	}

	if module != nil {
		cErr.module = *module
	}

	return cErr
}

// NewCustomError creates a new CustomError with the given code and error
func NewCustomError(code, message string, err error) CustomError {
	return CustomError{
		code:    code,
		message: message,
		err:     err,
	}
}

// ErrorPresenter is a custom error presenter for the GraphQL server
func ErrorPresenter(ctx context.Context, e error) *gqlerror.Error {
	err := graphql.DefaultErrorPresenter(ctx, e)

	var customError CustomErrorType
	switch e := e.(type) {
	case CustomErrorType:
		customError = e
	case *gqlerror.Error:
		var ok bool

		customError, ok = e.Err.(CustomErrorType)
		if !ok {
			return err
		}
	default:
		// default to the original error
		return err
	}

	// Add custom error code and message to the extensions
	if err.Extensions == nil {
		err.Extensions = make(map[string]interface{})
	}

	// add the code to the extensions
	if customError.Code() != "" {
		err.Extensions[ExtensionCodeKey] = customError.Code()
	}

	// add the message to the extensions if it is not empty
	if customError.Message() != "" {
		err.Extensions[ExtensionMessageKey] = customError.Message()
	}

	// add the module
	if customError.Module().String() != "" {
		err.Extensions[ExtensionModuleKey] = customError.Module()
	}

	return err
}
