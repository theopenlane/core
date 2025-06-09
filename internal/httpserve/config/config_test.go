package config

import (
	"testing"
)

func TestWithDefaultTLSConfig(t *testing.T) {
	t.Parallel()

	var c Config
	c = c.WithDefaultTLSConfig()
	if !c.Settings.Server.TLS.Enabled {
		t.Fatal("expected TLS enabled")
	}
	if c.Settings.Server.TLS.Config != DefaultTLSConfig {
		t.Fatal("expected default tls config set")
	}
}

func TestWithTLSCerts(t *testing.T) {
	t.Parallel()

	c := &Config{}
	c.WithTLSCerts("a", "b")
	if c.Settings.Server.TLS.CertFile != "a" || c.Settings.Server.TLS.CertKey != "b" {
		t.Fatal("certs not set")
	}
}

func TestWithAutoCert(t *testing.T) {
	t.Parallel()

	c := &Config{}
	c.WithAutoCert("example.com")
	if !c.Settings.Server.TLS.Enabled {
		t.Fatal("expected TLS enabled")
	}
	if c.Settings.Server.TLS.Config == nil || c.Settings.Server.TLS.Config.GetCertificate == nil {
		t.Fatal("expected cert manager configured")
	}
	want := []string{"acme-tls/1"}
	if len(c.Settings.Server.TLS.Config.NextProtos) == 0 || c.Settings.Server.TLS.Config.NextProtos[0] != want[0] {
		t.Fatal("NextProtos not set")
	}
}

func TestConfigGet(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	got, err := cfg.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != cfg {
		t.Fatal("expected same config returned")
	}
}
