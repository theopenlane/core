package rule

import "errors"

// ErrAdminOnlyField is returned when a user attempts to set a field that is only allowed to be set by an admin
var ErrAdminOnlyField = errors.New("invalid input: attempted to set admin only field")
