package hooks

import (
	"net/http"
	"time"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/favicon"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

const avatarFetchTimeout = 1 * time.Minute

var avatarDiscoveryClient = &http.Client{
	Timeout: avatarFetchTimeout,
}

// OrganizationAvatarListeners returns the organization avatar discovery listener
func OrganizationAvatarListeners() []gala.Registration {
	return []gala.Registration{entityops.MutationListener{
		Schema:     entityops.SchemaOrganization,
		Label:      "avatar",
		Operations: []string{entityops.OpCreate},
		Caller:     internalOperationBypassCaller,
		Handle:     handleOrganizationAvatarCreated,
	}}
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
		logx.FromContext(inv.Context).Err(err).Msg("organization avatar discovery failed")
		return nil
	}

	if avatarURL == "" {
		return nil
	}

	return inv.Client.Organization.UpdateOneID(inv.EntityID).
		SetAvatarRemoteURL(avatarURL).
		Exec(inv.Context)
}
