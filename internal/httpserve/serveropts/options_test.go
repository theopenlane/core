package serveropts

import (
	"os"
	"testing"

	coreconfig "github.com/theopenlane/core/config"
	serverconfig "github.com/theopenlane/core/internal/httpserve/config"
)

func TestWithGeneratedKeys(t *testing.T) {
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
