package contextx

import (
	"context"

	utilsctx "github.com/theopenlane/utils/contextx"
)

// ownershipTransferKey marks org membership role changes made by an authorized ownership transfer
var ownershipTransferKey = utilsctx.NewKey[bool]()

// WithOwnershipTransfer returns a new context that allows owner role changes on org memberships
// only set this on the role updates performed by an ownership transfer
func WithOwnershipTransfer(ctx context.Context) context.Context {
	return ownershipTransferKey.Set(ctx, true)
}

// IsOwnershipTransfer reports whether the context is performing an authorized ownership transfer
func IsOwnershipTransfer(ctx context.Context) bool {
	transfer, _ := ownershipTransferKey.Get(ctx)

	return transfer
}
