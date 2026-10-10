package registry

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// unreplaced returns the names of old not carried or replaced by next
func unreplaced[T any](old, next []T, name func(T) string, replaces func(T) []string) []string {
	carried := lo.FlatMap(next, func(entry T, _ int) []string {
		return append([]string{name(entry)}, replaces(entry)...)
	})

	return lo.Without(lo.Map(old, func(entry T, _ int) string { return name(entry) }), carried...)
}

// GateSurfaceChange refuses dropping a connection, operation, or webhook of the committed surface without a replacement
func GateSurfaceChange(old, next Surface) error {
	var refused []string

	refuse := func(kind string, names []string) {
		if len(names) > 0 {
			refused = append(refused, kind+" "+strings.Join(names, ", "))
		}
	}

	refuse("credential", unreplaced(old.Credentials, next.Credentials,
		func(credential SurfaceCredential) string { return credential.Ref },
		func(credential SurfaceCredential) []string { return credential.Replaces }))

	refuse("operation", unreplaced(old.Operations, next.Operations,
		func(operation SurfaceOperation) string { return operation.Name },
		func(operation SurfaceOperation) []string { return operation.Replaces }))

	refuse("webhook", unreplaced(old.Webhooks, next.Webhooks,
		func(webhook SurfaceWebhook) string { return webhook.Name },
		func(webhook SurfaceWebhook) []string { return webhook.Replaces }))

	if len(refused) == 0 {
		return nil
	}

	return fmt.Errorf("%w: %s: %s", ErrDestructiveSurfaceChange, next.ID, strings.Join(refused, "; "))
}
