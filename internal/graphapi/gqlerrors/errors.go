package gqlerrors

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const (
	// ExtensionCodeKey is the key for the error code in the extensions
	ExtensionCodeKey = "code"
	// ExtensionMessageKey is the key for the error message in the extensions
	ExtensionMessageKey = "message"
)

// CustomErrorType is an interface that defines a custom error type
type CustomErrorType interface {
	error
	Code() string
	Message() string
}

// CustomError is a struct that implements the CustomErrorType interface
type CustomError struct {
	code    string
	message string
	err     error
}

func (e CustomError) Error() string {
	return e.err.Error()
}

func (e CustomError) Code() string {
	return e.code
}

// Message returns the message of the error
func (e CustomError) Message() string {
	return e.message
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

	return err
}
