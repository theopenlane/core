package httpsling

import (
	"errors"
)

var (
	// ErrUnsupportedContentType is returned when the content type is unsupported
	ErrUnsupportedContentType = errors.New("unsupported content type")
	// ErrUnsupportedDataType is returned when the data type is unsupported
	ErrUnsupportedDataType = errors.New("unsupported data type")
	// ErrEncodingFailed is returned when the encoding fails
	ErrEncodingFailed = errors.New("encoding failed")
	// ErrRequestCreationFailed is returned when the request cannot be created
	ErrRequestCreationFailed = errors.New("failed to create request")
	// ErrResponseReadFailed is returned when the response cannot be read
	ErrResponseReadFailed = errors.New("failed to read response")
	// ErrUnsupportedScheme is returned when the proxy scheme is unsupported
	ErrUnsupportedScheme = errors.New("unsupported proxy scheme")
	// ErrUnsupportedFormFieldsType is returned when the form fields type is unsupported
	ErrUnsupportedFormFieldsType = errors.New("unsupported form fields type")
	// ErrNotSupportSaveMethod is returned when the provided type for saving is not supported
	ErrNotSupportSaveMethod = errors.New("the provided type for saving is not supported")
	// ErrInvalidTransportType is returned when the transport type is invalid
	ErrInvalidTransportType = errors.New("invalid transport type")
	// ErrResponseNil is returned when the response is nil
	ErrResponseNil = errors.New("response is nil")
	// ErrFailedToCloseResponseBody is returned when the response body cannot be closed
	ErrFailedToCloseResponseBody = errors.New("failed to close response body")
	// ErrMapper
	ErrMapper = "%w: %v"
)
