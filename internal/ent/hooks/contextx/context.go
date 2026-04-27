package contextx

import "context"

// SkipCustomEnumDeleteKey is the context key used to skip the "in use" errors check during enum deletion.
// This is used during organization cascade deletion where the deletion order is handled by EdgeCleanup.
// else the custom deletion by default will check if the enum is being used by another other object.
// But with this, we can just skip the check because when the org itself is deleted, it cascades to delete the
// custom enums too
type SkipCustomEnumDeleteKey string

const (
	// SkipCustomEnumInUseCheck is the context value that triggers skipping the "in use" check/error during enum deletion.
	SkipCustomEnumInUseCheck SkipCustomEnumDeleteKey = "custom_enum_cascade_delete_operation"
)

// WithSkipEnumInUseCheck returns a new context with the skip flag set for custom enums deletion.
// This should be used when deleting CustomTypeEnums as part of a cascade delete
// where the deletion order is handled by EdgeCleanup.
func WithSkipEnumInUseCheck(ctx context.Context) context.Context {
	return context.WithValue(ctx, SkipCustomEnumInUseCheck, true)
}
