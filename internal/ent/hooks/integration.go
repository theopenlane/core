package hooks

import (
	"context"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	integrationtypes "github.com/theopenlane/core/internal/integrations/types"
)

// HookIntegrationPrimaryDirectory enforces the one-primary-directory-per-org invariant
// When an integration is set as the primary directory, all sibling integrations in the
// same organization have their primary_directory flag cleared
func HookIntegrationPrimaryDirectory() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.IntegrationFunc(func(ctx context.Context, m *generated.IntegrationMutation) (generated.Value, error) {
			if shouldSkipIntegrationPrimaryDirectorySync(ctx) {
				return next.Mutate(ctx, m)
			}

			primaryDirectory, ok := m.PrimaryDirectory()
			if !ok || !primaryDirectory {
				return next.Mutate(ctx, m)
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			inst := retVal.(*generated.Integration)

			return retVal, m.Client().Integration.Update().
				Where(
					integration.OwnerID(inst.OwnerID),
					integration.IDNEQ(inst.ID),
					integration.PrimaryDirectory(true),
				).
				SetPrimaryDirectory(false).
				Exec(privacy.DecisionContext(withSkipIntegrationPrimaryDirectorySync(ctx), privacy.Allow))
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// shouldSkipIntegrationPrimaryDirectorySync reports whether sibling clearing should be bypassed
func shouldSkipIntegrationPrimaryDirectorySync(ctx context.Context) bool {
	metadata, _ := integrationtypes.ExecutionMetadataFromContext(ctx)
	return metadata.SkipPrimaryDirectorySync
}

// withSkipIntegrationPrimaryDirectorySync marks a context to bypass recursive hook re-entry
func withSkipIntegrationPrimaryDirectorySync(ctx context.Context) context.Context {
	metadata, _ := integrationtypes.ExecutionMetadataFromContext(ctx)
	metadata.SkipPrimaryDirectorySync = true

	return integrationtypes.WithExecutionMetadata(ctx, metadata)
}

// HookIntegrationCampaignEmail enforces the one-campaign-email-per-org invariant.
// When an integration is flagged as the campaign email provider, all sibling
// integrations in the same organization have their campaign_email flag cleared
func HookIntegrationCampaignEmail() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.IntegrationFunc(func(ctx context.Context, m *generated.IntegrationMutation) (generated.Value, error) {
			if shouldSkipIntegrationCampaignEmailSync(ctx) {
				return next.Mutate(ctx, m)
			}

			campaignEmail, ok := m.CampaignEmail()
			if !ok || !campaignEmail {
				return next.Mutate(ctx, m)
			}

			retVal, err := next.Mutate(ctx, m)
			if err != nil {
				return nil, err
			}

			inst := retVal.(*generated.Integration)

			return retVal, m.Client().Integration.Update().
				Where(
					integration.OwnerID(inst.OwnerID),
					integration.IDNEQ(inst.ID),
					integration.CampaignEmail(true),
				).
				SetCampaignEmail(false).
				Exec(privacy.DecisionContext(withSkipIntegrationCampaignEmailSync(ctx), privacy.Allow))
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// shouldSkipIntegrationCampaignEmailSync reports whether sibling clearing should be bypassed
func shouldSkipIntegrationCampaignEmailSync(ctx context.Context) bool {
	metadata, _ := integrationtypes.ExecutionMetadataFromContext(ctx)
	return metadata.SkipCampaignEmailSync
}

// withSkipIntegrationCampaignEmailSync marks a context to bypass recursive hook re-entry
func withSkipIntegrationCampaignEmailSync(ctx context.Context) context.Context {
	metadata, _ := integrationtypes.ExecutionMetadataFromContext(ctx)
	metadata.SkipCampaignEmailSync = true

	return integrationtypes.WithExecutionMetadata(ctx, metadata)
}
