package serveropts

import (
	"os"
	"testing"

	"github.com/theopenlane/iam/sessions"

	coreconfig "github.com/theopenlane/core/config"
	serverconfig "github.com/theopenlane/core/internal/httpserve/config"
	"github.com/theopenlane/core/internal/httpserve/handlers"
)

func TestWithGeneratedKeys(t *testing.T) {
	t.Parallel()

	defer os.Remove("private_key.pem")
	so := &ServerOptions{Config: serverconfig.Config{Settings: coreconfig.Config{}}}
	opt := WithGeneratedKeys()
	opt.apply(so)
	if _, err := os.Stat("private_key.pem"); err != nil {
		t.Fatalf("expected key file created: %v", err)
	}
	if len(so.Config.Settings.Auth.Token.Keys) == 0 {
		t.Fatalf("expected keys map to be populated")
	}
}

func TestWithAuth_DisabledIntegrationRegistry(t *testing.T) {
	t.Parallel()

	so := &ServerOptions{
		Config: serverconfig.Config{
			Settings: coreconfig.Config{
				IntegrationOauthProvider: handlers.IntegrationOauthProviderConfig{
					Enabled: false,
				},
			},
			SessionConfig: &sessions.SessionConfig{},
		},
	}

	WithAuth().apply(so)

	if so.Config.Handler.IntegrationRegistry != nil {
		t.Fatalf("expected integration registry to remain nil when disabled")
	}
}

func TestWithAuth_EnabledIntegrationRegistry(t *testing.T) {
	t.Parallel()

	so := &ServerOptions{
		Config: serverconfig.Config{
			Settings: coreconfig.Config{
				IntegrationOauthProvider: handlers.IntegrationOauthProviderConfig{
					Enabled: true,
				},
			},
			SessionConfig: &sessions.SessionConfig{},
		},
	}

	WithAuth().apply(so)

	if so.Config.Handler.IntegrationRegistry == nil {
		t.Fatalf("expected integration registry to be initialized when enabled")
	}
}

// TestWithAuth_EnabledIntegrationRegistry_GitHubApp ensures GitHub App settings enable the registry.
func TestWithAuth_EnabledIntegrationRegistry_GitHubApp(t *testing.T) {
	t.Parallel()

	so := &ServerOptions{
		Config: serverconfig.Config{
			Settings: coreconfig.Config{
				IntegrationGitHubApp: handlers.IntegrationGitHubAppConfig{
					Enabled: true,
				},
			},
			SessionConfig: &sessions.SessionConfig{},
		},
	}

	WithAuth().apply(so)

	if so.Config.Handler.IntegrationRegistry == nil {
		t.Fatalf("expected integration registry to be initialized when GitHub App integration enabled")
	}
}
