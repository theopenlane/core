//go:build examples

package config

import (
	"os"
	"testing"
)

func TestEnvOverride(t *testing.T) {
	os.Setenv("OBJECTS_EXAMPLES_OPENLANE_BASEURL", "http://overridden.example.com")
	defer os.Unsetenv("OBJECTS_EXAMPLES_OPENLANE_BASEURL")

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	expected := "http://overridden.example.com"
	if cfg.Openlane.BaseURL != expected {
		t.Errorf("Expected BaseURL %s, got %s", expected, cfg.Openlane.BaseURL)
	}
}

func TestEnvOverrideMultiple(t *testing.T) {
	os.Setenv("OBJECTS_EXAMPLES_OPENLANE_BASEURL", "http://env.example.com")
	os.Setenv("OBJECTS_EXAMPLES_OPENLANE_EMAIL", "env@example.com")
	os.Setenv("OBJECTS_EXAMPLES_OPENLANE_ORGANIZATIONID", "env-org-123")
	defer func() {
		os.Unsetenv("OBJECTS_EXAMPLES_OPENLANE_BASEURL")
		os.Unsetenv("OBJECTS_EXAMPLES_OPENLANE_EMAIL")
		os.Unsetenv("OBJECTS_EXAMPLES_OPENLANE_ORGANIZATIONID")
	}()

	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Openlane.BaseURL != "http://env.example.com" {
		t.Errorf("Expected BaseURL http://env.example.com, got %s", cfg.Openlane.BaseURL)
	}

	if cfg.Openlane.Email != "env@example.com" {
		t.Errorf("Expected Email env@example.com, got %s", cfg.Openlane.Email)
	}

	if cfg.Openlane.OrganizationID != "env-org-123" {
		t.Errorf("Expected OrganizationID env-org-123, got %s", cfg.Openlane.OrganizationID)
	}
}
