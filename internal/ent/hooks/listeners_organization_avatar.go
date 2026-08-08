package hooks

import (
	"context"
	"net/http"
	"time"

	"entgo.io/ent"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/favicon"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

const avatarFetchTimeout = 1 * time.Minute

var avatarDiscoveryClient = &http.Client{
	Timeout: avatarFetchTimeout,
}

// RegisterGalaOrganizationAvatarListeners registers organization avatar discovery on Gala.
func RegisterGalaOrganizationAvatarListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return registerMutationListeners(g, entityops.MutationListener{
		Schema:     generated.TypeOrganization,
		Label:      "avatar",
		Operations: []string{ent.OpCreate.String()},
		Elevate: func(ctx context.Context, _ entityops.MutationPayload) context.Context {
			return privacy.DecisionContext(rule.WithInternalContext(ctx), privacy.Allow)
		},
		Handle: handleOrganizationAvatarCreated,
	})
}

// handleOrganizationAvatarCreated fetches icons from the domain name instead and sets it as the remote logo url
func handleOrganizationAvatarCreated(inv entityops.Invocation, _ entityops.MutationPayload) error {
	org, err := inv.Client.Organization.Query().
		Where(organization.IDEQ(inv.EntityID)).
		WithSetting().
		Only(inv.Context)
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

	avatarURL, err := favicon.Discover(inv.Context, avatarDiscoveryClient, setting.Domains)
	if err != nil {
		logx.FromContext(inv.Context).Err(err).Str("organization_id", inv.EntityID).Msg("organization avatar discovery failed")
		return nil
	}

	if avatarURL == "" {
		return nil
	}

	return inv.Client.Organization.UpdateOneID(inv.EntityID).
		SetAvatarRemoteURL(avatarURL).
		Exec(inv.Context)
}
