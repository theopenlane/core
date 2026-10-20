package azuresecuritycenter

import (
	"context"
	"fmt"

	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/logx"
)

const (
	// lighthouseManualRemovalMessage tells the customer how to revoke access when Openlane could not remove its delegations
	lighthouseManualRemovalMessage = "Openlane could not remove all of its Lighthouse delegations. Remove the Openlane Defender for Cloud offer from Service providers in the Azure portal for each delegated subscription."
	// lighthouseNoneRemovedMessage is shown when no delegation carried the installation's connection code
	lighthouseNoneRemovedMessage = "No Openlane Lighthouse delegations were found to remove."
	// lighthouseRemovedMessageFormat reports how many subscriptions had their delegation removed
	lighthouseRemovedMessageFormat = "Removed the Openlane Lighthouse delegation from %d subscription(s)."
)

// lighthouseDisconnectDetails lists the subscriptions whose Lighthouse delegation was removed on disconnect
type lighthouseDisconnectDetails struct {
	// RemovedSubscriptionIDs are the subscriptions Openlane no longer has access to
	RemovedSubscriptionIDs []string `json:"removedSubscriptionIds,omitempty"`
}

// disconnectLighthouse removes the installation's Lighthouse delegations, falling back to manual removal instructions on failure
func disconnectLighthouse(cfg Config) func(context.Context, types.DisconnectRequest[lighthouseCred]) (types.DisconnectResult, error) {
	return func(ctx context.Context, req types.DisconnectRequest[lighthouseCred]) (types.DisconnectResult, error) {
		client, err := lighthouseClientBuilder(cfg)(ctx, types.ConnectionRequest[lighthouseCred]{Integration: req.Integration, Credential: req.Credential})
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("azuresecuritycenter: failed to build lighthouse client for disconnect")
			return types.DisconnectResult{Message: lighthouseManualRemovalMessage}, nil
		}

		removed, err := client.removeDelegations(ctx)

		details, encodeErr := jsonx.ToRawMessage(lighthouseDisconnectDetails{RemovedSubscriptionIDs: removed})
		if encodeErr != nil {
			return types.DisconnectResult{}, ErrDisconnectDetailsEncode
		}

		switch {
		case err != nil:
			logx.FromContext(ctx).Error().Err(err).Int("removed", len(removed)).Msg("azuresecuritycenter: failed to remove lighthouse delegations")
			return types.DisconnectResult{Message: lighthouseManualRemovalMessage, Details: details}, nil
		case len(removed) == 0:
			return types.DisconnectResult{Message: lighthouseNoneRemovedMessage, Details: details}, nil
		default:
			return types.DisconnectResult{Message: fmt.Sprintf(lighthouseRemovedMessageFormat, len(removed)), Details: details}, nil
		}
	}
}

// removeDelegations deletes the installation's Lighthouse delegation from every tenant subscription that carries it, returning the subscriptions removed so far
func (c *Client) removeDelegations(ctx context.Context) ([]string, error) {
	ids, err := c.arm.ListSubscriptionIDs(ctx, c.tenantID)
	if err != nil {
		return nil, err
	}

	var removed []string

	for _, id := range ids {
		delegation, ok, err := c.findDelegation(ctx, id)
		if err != nil {
			return removed, err
		}

		if !ok {
			continue
		}

		if err := c.arm.DeleteDelegation(ctx, delegation.AssignmentID); err != nil {
			return removed, err
		}

		removed = append(removed, id)
	}

	return removed, nil
}
