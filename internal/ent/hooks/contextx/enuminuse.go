package contextx

import (
	"context"

	utilsctx "github.com/theopenlane/utils/contextx"
)

// skipEnumInUseCheckKey skips the "in use" check during custom enum deletion.
// This is used during organization cascade deletion where the deletion order is handled by
// EdgeCleanup, the custom enum deletion would otherwise check whether the enum is still used by
// another object. When the organization itself is deleted the cascade removes those objects too,
// so the check has nothing useful to say
var skipEnumInUseCheckKey = utilsctx.NewKey[bool]()

// WithSkipEnumInUseCheck returns a new context with the skip flag set for custom enums deletion.
// This should be used when deleting CustomTypeEnums as part of a cascade delete
// where the deletion order is handled by EdgeCleanup
func WithSkipEnumInUseCheck(ctx context.Context) context.Context {
	return skipEnumInUseCheckKey.Set(ctx, true)
}

// SkipEnumInUseCheckEnabled reports whether the custom enum "in use" check should be skipped
func SkipEnumInUseCheckEnabled(ctx context.Context) bool {
	skip, _ := skipEnumInUseCheckKey.Get(ctx)

	return skip
}
