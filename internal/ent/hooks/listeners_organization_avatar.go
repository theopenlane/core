package hooks

import (
	"time"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/favicon"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/urlx"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"
	"github.com/theopenlane/logx"
)

// init registers the organization avatar listeners with their default options so gala
// setup picks them up automatically
func init() { registerListeners(func() []gala.Registration { return OrganizationAvatarListeners() }) }

const avatarFetchTimeout = 5 * time.Second

// OrganizationAvatarListenerOption customizes avatar discovery
type OrganizationAvatarListenerOption func(*organizationAvatarListenerConfig)

type organizationAvatarListenerConfig struct {
	requester *httpsling.Requester
}

// WithOrganizationAvatarRequester sets the avatar discovery requester
func WithOrganizationAvatarRequester(requester *httpsling.Requester) OrganizationAvatarListenerOption {
	return func(config *organizationAvatarListenerConfig) {
		if requester != nil {
			config.requester = requester
		}
	}
}

// OrganizationAvatarListeners discovers an avatar from the organization's domains after creation
func OrganizationAvatarListeners(opts ...OrganizationAvatarListenerOption) []gala.Registration {
	config := &organizationAvatarListenerConfig{}
	for _, opt := range opts {
		opt(config)
	}

	return []gala.Registration{entityops.MutationListener{
		Schema:     entityops.SchemaOrganization,
		Operations: []string{entityops.OpCreate},
		Caller:     internalOperationBypassCaller,
		Handle: func(inv entityops.Invocation, payload entityops.MutationPayload) error {
			return handleOrganizationAvatarCreated(inv, payload, config.requester)
		},
	}}
}

// handleOrganizationAvatarCreated fetches icons from the domain name instead and sets it as the remote logo url
func handleOrganizationAvatarCreated(inv entityops.Invocation, _ entityops.MutationPayload, requester *httpsling.Requester) error {
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

	if requester == nil {
		requester, err = urlx.NewRequester(httpsling.Client(httpclient.Timeout(avatarFetchTimeout)))
		if err != nil {
			return err
		}
	}

	avatarURL, err := favicon.Discover(inv.Context, requester, setting.Domains)
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
