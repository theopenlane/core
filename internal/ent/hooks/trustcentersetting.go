package hooks

import (
	"context"
	"strings"
	"time"

	"entgo.io/ent"
	"github.com/microcosm-cc/bluemonday"

	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/hook"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/trustcenterurl"
	"github.com/theopenlane/logx"
)

var trustCenterConfig TrustCenterConfig

// SetTrustCenterConfig sets the trust center configuration. It also populates the shared
// trustcenterurl config so the integration runtime can build trust center links without importing
// this package
func SetTrustCenterConfig(cfg TrustCenterConfig) {
	trustCenterConfig = cfg

	trustcenterurl.SetConfig(trustcenterurl.Config{
		CnameTarget:              cfg.CnameTarget,
		DefaultTrustCenterDomain: cfg.DefaultTrustCenterDomain,
		CacheRefreshScheme:       cfg.CacheRefreshScheme,
	})
}

// TrustCenterConfig holds the trust center configuration
type TrustCenterConfig struct {
	// CnameTarget is the CNAME target custom trust center domains point at
	CnameTarget string
	// PreviewZoneID is the cloudflare zone id that holds trust center preview domain records
	PreviewZoneID string
	// DefaultTrustCenterDomain is the shared domain slugged trust centers are served from
	DefaultTrustCenterDomain string
	// CacheRefreshScheme is the URL scheme used for cache refresh requests (http in tests, https otherwise)
	CacheRefreshScheme string
	// CacheRefreshSecret is the shared secret sent on cache refresh requests so the trust center
	// frontend can authenticate the forced-fresh render; no header is sent when empty
	CacheRefreshSecret string
}

// trustCenterOverviewPolicy is the HTML sanitization policy applied to the trust center overview,
// which the public trust center frontend renders as raw HTML
var trustCenterOverviewPolicy = bluemonday.UGCPolicy()

// HookTrustCenterSettingSanitize strips unsafe HTML from the trust center overview on create and
// update so a tenant editor cannot store script the public trust center would execute
func HookTrustCenterSettingSanitize() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterSettingFunc(func(ctx context.Context, m *generated.TrustCenterSettingMutation) (generated.Value, error) {
			if overview, ok := m.Overview(); ok {
				m.SetOverview(trustCenterOverviewPolicy.Sanitize(overview))
			}

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne|ent.OpUpdate)
}

// HookTrustCenterSetting process files for trust center settings
func HookTrustCenterSetting() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.TrustCenterSettingFunc(func(ctx context.Context, m *generated.TrustCenterSettingMutation) (generated.Value, error) {
			logx.FromContext(ctx).Debug().Msg("trust center setting hook")

			if err := setDefaultCompanyName(ctx, m); err != nil {
				return nil, err
			}

			// set email to lowercase if provided
			email, ok := m.SecurityContact()
			if ok {
				m.SetSecurityContact(strings.ToLower(email))
			}

			initializeSubprocessorWatermark(ctx, m)

			return next.Mutate(ctx, m)
		})
	}, ent.OpCreate|ent.OpUpdateOne)
}

// initializeSubprocessorWatermark stamps the subprocessor notification watermark when the
// notify-on-subprocessor-change flag turns on, so subscribers are only notified about changes made
// after opting in rather than the trust center's entire subprocessor history. A watermark set
// explicitly on the same mutation wins
func initializeSubprocessorWatermark(ctx context.Context, m *generated.TrustCenterSettingMutation) {
	enabled, ok := m.NotifySubscribersOnSubprocessorChange()
	if !ok || !enabled {
		return
	}

	if _, ok := m.SubprocessorsNotifiedAt(); ok {
		return
	}

	if m.Op().Is(ent.OpUpdateOne) {
		wasEnabled, err := m.OldNotifySubscribersOnSubprocessorChange(ctx)
		if err == nil && wasEnabled {
			return
		}
	}

	m.SetSubprocessorsNotifiedAt(time.Now())
}

// setDefaultCompanyName will add the company name either from the org display name or if provided in the mutation
func setDefaultCompanyName(ctx context.Context, m *generated.TrustCenterSettingMutation) error {

	if m.Op().Is(ent.OpUpdateOne) {
		oldName, err := m.OldCompanyName(ctx)
		if err == nil && oldName != "" {
			return nil
		}
	}

	name, ok := m.CompanyName()
	if ok && name != "" {
		return nil
	}

	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return err
	}

	org, err := m.Client().Organization.
		Query().Select(organization.FieldDisplayName).
		Where(organization.ID(orgID)).
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("owner_id", orgID).Msg("failed to get organization for company name default")
		return err
	}

	m.SetCompanyName(org.DisplayName)
	return nil
}
