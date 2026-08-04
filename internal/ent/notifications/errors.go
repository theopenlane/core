package notifications

import "errors"

// ErrFailedToGetClient is returned when the client cannot be retrieved from context
var ErrFailedToGetClient = errors.New("failed to get client from context")
