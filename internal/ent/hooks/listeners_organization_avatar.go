package hooks

import (
	"net/http"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/favicon"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

const avatarFetchTimeout = 1 * time.Minute

var avatarDiscoveryClient = &http.Client{
	Timeout: avatarFetchTimeout,
}

// RegisterGalaOrganizationAvatarListeners registers organization avatar discovery on Gala.
func RegisterGalaOrganizationAvatarListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry, gala.Definition[eventqueue.MutationGalaPayload]{
		Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeOrganization),
		Name:       "organization.avatar",
		Operations: []string{ent.OpCreate.String()},
		Handle:     handleOrganizationAvatarCreated,
	})
}

// handleOrganizationAvatarCreated fetches icons from the domain name instead and sets it as the remote logo url
func handleOrganizationAvatarCreated(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	orgID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || orgID == "" {
		return nil
	}

	allowCtx := workflows.AllowContext(ctx.Context)

	org, err := client.Organization.Query().
		Where(organization.IDEQ(orgID)).
		WithSetting().
		Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	setting, err := org.Edges.SettingOrErr()
	if err != nil || setting == nil || len(setting.Domains) == 0 {
		return nil
	}

	avatarURL, err := favicon.Discover(ctx.Context, avatarDiscoveryClient, setting.Domains)
	if err != nil {
		logx.FromContext(ctx.Context).Err(err).
			Str("organization_id", orgID).
			Msg("organization avatar discovery failed")
		return nil
	}

	if avatarURL == "" {
		return nil
	}

	return client.Organization.UpdateOneID(orgID).
		SetAvatarRemoteURL(avatarURL).
		Exec(allowCtx)
}
