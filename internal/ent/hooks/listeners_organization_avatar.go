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
func RegisterGalaOrganizationAvatarListeners(g *gala.Gala) ([]gala.ListenerID, error) {
	return eventqueue.RegisterMutationListeners(g, eventqueue.MutationListener{
		Schema:     generated.TypeOrganization,
		Name:       "organization.avatar",
		Operations: []string{ent.OpCreate.String()},
		Handle:     handleOrganizationAvatarCreated,
	})
}

// handleOrganizationAvatarCreated fetches icons from the domain name instead and sets it as the remote logo url
func handleOrganizationAvatarCreated(inv eventqueue.Invocation, _ eventqueue.MutationGalaPayload) error {
	allowCtx := workflows.AllowContext(inv.Context)

	org, err := inv.Client.Organization.Query().
		Where(organization.IDEQ(inv.EntityID)).
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

	avatarURL, err := favicon.Discover(inv.Context, avatarDiscoveryClient, setting.Domains)
	if err != nil {
		logx.FromContext(inv.Context).Err(err).
			Str("organization_id", inv.EntityID).
			Msg("organization avatar discovery failed")
		return nil
	}

	if avatarURL == "" {
		return nil
	}

	return inv.Client.Organization.UpdateOneID(inv.EntityID).
		SetAvatarRemoteURL(avatarURL).
		Exec(allowCtx)
}
