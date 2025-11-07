//go:build cli

package orgsubscription

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/theopenlane/utils/cli/tables"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func fetchOrgSubscriptions(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}

	id := cmd.Config.String("id")
	if id != "" {
		return client.GetOrgSubscriptionByID(ctx, id)
	}

	return client.GetAllOrgSubscriptions(ctx)
}

func renderOrgSubscriptions(result any) error {
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return speccli.PrintJSON(result)
	}

	subs, err := collectOrgSubscriptions(result)
	if err != nil {
		return err
	}

	emitOrgSubscriptionTable(subs)
	return nil
}

func emitOrgSubscriptionTable(subs []openlaneclient.OrgSubscription) {
	writer := tables.NewTableWriter(cmd.RootCmd.OutOrStdout(), "ID", "Active", "StripeSubscriptionStatus", "ExpiresAt")

	for _, sub := range subs {
		expires := ""
		if sub.ExpiresAt != nil {
			expires = sub.ExpiresAt.String()
		} else if sub.TrialExpiresAt != nil {
			expires = sub.TrialExpiresAt.String()
		}

		status := ""
		if sub.StripeSubscriptionStatus != nil {
			status = *sub.StripeSubscriptionStatus
		}

		writer.AddRow(sub.ID, sub.Active, status, expires)
	}

	writer.Render()
}

func collectOrgSubscriptions(result any) ([]openlaneclient.OrgSubscription, error) {
	var nodes []any

	switch v := result.(type) {
	case *openlaneclient.GetAllOrgSubscriptions:
		for _, edge := range v.OrgSubscriptions.Edges {
			if edge != nil && edge.Node != nil {
				nodes = append(nodes, edge.Node)
			}
		}
	case *openlaneclient.GetOrgSubscriptions:
		for _, edge := range v.OrgSubscriptions.Edges {
			if edge != nil && edge.Node != nil {
				nodes = append(nodes, edge.Node)
			}
		}
	case *openlaneclient.GetOrgSubscriptionByID:
		if v != nil && v.GetOrgSubscription() != nil {
			nodes = append(nodes, v.GetOrgSubscription())
		}
	}

	if len(nodes) == 0 {
		payload, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}

		var list []openlaneclient.OrgSubscription
		if err := json.Unmarshal(payload, &list); err != nil {
			return nil, err
		}
		return list, nil
	}

	payload, err := json.Marshal(nodes)
	if err != nil {
		return nil, err
	}

	var list []openlaneclient.OrgSubscription
	if err := json.Unmarshal(payload, &list); err != nil {
		return nil, err
	}

	return list, nil
}
