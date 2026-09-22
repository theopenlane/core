package vendorenrich

import "errors"

// ErrVendorEntityTypeMissing indicates the organization has no vendor entity type
var ErrVendorEntityTypeMissing = errors.New("vendorenrich: vendor entity type missing")
