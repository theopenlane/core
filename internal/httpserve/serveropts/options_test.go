package serveropts

import (
	"os"
	"testing"

	"github.com/theopenlane/iam/sessions"

	coreconfig "github.com/theopenlane/core/config"
	serverconfig "github.com/theopenlane/core/internal/httpserve/config"
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

func TestWithIntegrationsRuntime_NilDB(t *testing.T) {
	t.Parallel()

	so := &ServerOptions{
		Config: serverconfig.Config{
			Settings:      coreconfig.Config{},
			SessionConfig: &sessions.SessionConfig{},
		},
	}

	WithIntegrationsRuntime(nil).apply(so)

	if so.Config.Handler.IntegrationsRuntime != nil {
		t.Fatalf("expected integrations runtime to remain nil when DB is nil")
	}
}
