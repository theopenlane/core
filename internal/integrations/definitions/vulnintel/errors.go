package vulnintel

import "errors"

// ErrOSVRequest indicates an OSV lookup failed or returned an unexpected response
var ErrOSVRequest = errors.New("vulnintel: osv lookup failed")
