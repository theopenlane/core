package serveropts

import (
	"github.com/rs/zerolog/log"

	ent "github.com/theopenlane/core/internal/ent/generated"
	integrationruntime "github.com/theopenlane/core/internal/integrations/runtime"
	"github.com/theopenlane/core/internal/workflows/engine"
)

// WithIntegrationRuntime builds the integrations runtime from server settings and wires it
// into the handler. When a workflow engine is present it also injects integration dependencies.
func WithIntegrationRuntime(dbClient *ent.Client) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if dbClient == nil {
			return
		}

		ghApp := s.Config.Settings.IntegrationGitHubApp
		oauth := s.Config.Settings.IntegrationOauthProvider

		rt, err := integrationruntime.New(integrationruntime.Config{
			ProviderSpecs: s.Config.Settings.IntegrationProviders,
			DB:            dbClient,
			GitHubApp: integrationruntime.GitHubAppConfig{
				Enabled:            ghApp.Enabled,
				AppID:              ghApp.AppID,
				AppSlug:            ghApp.AppSlug,
				PrivateKey:         ghApp.PrivateKey,
				WebhookSecret:      ghApp.WebhookSecret,
				SuccessRedirectURL: ghApp.SuccessRedirectURL,
			},
			OAuth: integrationruntime.OAuthConfig{
				Enabled:            oauth.Enabled,
				SuccessRedirectURL: oauth.SuccessRedirectURL,
			},
		})
		if err != nil {
			log.Panic().Err(err).Msg("failed to initialize integration runtime")
		}

		s.Config.Handler.IntegrationRuntime = rt

		wf := s.Config.Handler.WorkflowEngine
		if wf == nil {
			return
		}

		if err := wf.SetIntegrationDeps(engine.IntegrationDeps{
			Registry:     rt.Registry(),
			Store:        rt.Store(),
			Operations:   rt.Operations(),
			MappingIndex: rt.Registry(),
		}); err != nil {
			log.Panic().Err(err).Msg("failed to wire integration deps into workflow engine")
		}
	})
}
