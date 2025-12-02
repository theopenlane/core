package gqlerrors

// Error code constants
const (
	// NotFoundErrorCode is the error code for not found errors
	NotFoundErrorCode = "NOT_FOUND"
	// ValidationErrorCode is the error code for validation errors
	ValidationErrorCode = "VALIDATION_ERROR"
	// ConflictErrorCode is the error code for conflict errors
	ConflictErrorCode = "CONFLICT"
	// InternalServerErrorCode is the error code for internal server errors
	InternalServerErrorCode = "INTERNAL_SERVER_ERROR"
	// UnauthorizedErrorCode is the error code for unauthorized errors
	UnauthorizedErrorCode = "UNAUTHORIZED"
	// AlreadyExistsErrorCode is the error code for already exists errors
	AlreadyExistsErrorCode = "ALREADY_EXISTS"
	// MaxAttemptsErrorCode is the error code for max attempts errors
	MaxAttemptsErrorCode = "MAX_ATTEMPTS"
	// BadRequestErrorCode is the error code for bad request errors
	BadRequestErrorCode = "BAD_REQUEST"
	// NoAccessToModule is the error code for when an org has no access to a
	// specific schema and module
	NoAccessToModule = "MODULE_NO_ACCESS"
)
