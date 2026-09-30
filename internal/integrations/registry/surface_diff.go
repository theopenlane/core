package registry

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samber/lo"
)

// unreplaced returns the retired names from old that no entry in next carries or replaces
func unreplaced[T any](old []string, next []T, name func(T) string, replaces func(T) []string) []string {
	carried := lo.FlatMap(next, func(entry T, _ int) []string {
		return append([]string{name(entry)}, replaces(entry)...)
	})

	return lo.Without(old, carried...)
}

// GateSurfaceChange refuses a next surface that drops a credential slot, operation or webhook the committed snapshot had without a registration replacing it
func GateSurfaceChange(target string, existing []byte, next Surface) error {
	var old Surface
	if err := json.Unmarshal(existing, &old); err != nil {
		return fmt.Errorf("decode %s: %w", target, err)
	}

	var refused []string

	refuse := func(kind string, names []string) {
		if len(names) > 0 {
			refused = append(refused, kind+" "+strings.Join(names, ", "))
		}
	}

	credentialRef := func(credential SurfaceCredential) string { return credential.Ref }
	operationName := func(operation SurfaceOperation) string { return operation.Name }
	webhookName := func(webhook SurfaceWebhook) string { return webhook.Name }

	refuse("credential", unreplaced(lo.Map(old.Credentials, func(credential SurfaceCredential, _ int) string { return credentialRef(credential) }), next.Credentials, credentialRef,
		func(credential SurfaceCredential) []string { return credential.Replaces }))

	refuse("operation", unreplaced(lo.Map(old.Operations, func(operation SurfaceOperation, _ int) string { return operationName(operation) }), next.Operations, operationName,
		func(operation SurfaceOperation) []string { return operation.Replaces }))

	refuse("webhook", unreplaced(lo.Map(old.Webhooks, func(webhook SurfaceWebhook, _ int) string { return webhookName(webhook) }), next.Webhooks, webhookName,
		func(webhook SurfaceWebhook) []string { return webhook.Replaces }))

	if len(refused) == 0 {
		return nil
	}

	return fmt.Errorf("%w: %s: %s", ErrDestructiveSurfaceChange, next.ID, strings.Join(refused, "; "))
}
