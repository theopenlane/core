//go:build examples

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Config is nil")
	}

	if cfg.Openlane.BaseURL == "" {
		t.Error("BaseURL should have a default value")
	}

	t.Logf("Loaded config: BaseURL=%s, Email=%s", cfg.Openlane.BaseURL, cfg.Openlane.Email)
}

func TestLoadWithExplicitPath(t *testing.T) {
	examplesDir := getExamplesDir()
	configPath := filepath.Join(examplesDir, defaultConfigFilePath)

	cfg, err := Load(&configPath)
	if err != nil {
		t.Fatalf("Load with explicit path failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Config is nil")
	}

	t.Logf("Config path: %s", configPath)
	t.Logf("Loaded config: BaseURL=%s", cfg.Openlane.BaseURL)
}

func TestNew(t *testing.T) {
	cfg := New()
	if cfg == nil {
		t.Fatal("New() returned nil")
	}

	if cfg.Openlane.BaseURL != "http://localhost:17608" {
		t.Errorf("Expected default BaseURL http://localhost:17608, got %s", cfg.Openlane.BaseURL)
	}
}

func TestConfigExists(t *testing.T) {
	exists := ConfigExists()
	t.Logf("Config exists: %v", exists)
}

func TestSaveOpenlaneConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	testDir := filepath.Join(tmpDir, "pkg/objects/examples")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	testConfig := &OpenlaneConfig{
		BaseURL:        "http://test.example.com",
		Email:          "test@example.com",
		Password:       "testpass",
		Token:          "testtoken",
		OrganizationID: "test-org",
		PAT:            "test-pat",
	}

	if err := SaveOpenlaneConfig(testConfig); err != nil {
		t.Fatalf("SaveOpenlaneConfig failed: %v", err)
	}

	if !ConfigExists() {
		t.Fatal("Config file should exist after saving")
	}

	configPath := filepath.Join(testDir, defaultConfigFilePath)
	loaded, err := Load(&configPath)
	if err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loaded.Openlane.BaseURL != testConfig.BaseURL {
		t.Errorf("Expected BaseURL %s, got %s", testConfig.BaseURL, loaded.Openlane.BaseURL)
	}

	if loaded.Openlane.Email != testConfig.Email {
		t.Errorf("Expected Email %s, got %s", testConfig.Email, loaded.Openlane.Email)
	}
}

func TestDeleteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	testDir := filepath.Join(tmpDir, "pkg/objects/examples")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	testConfig := &OpenlaneConfig{
		BaseURL:        "http://test.example.com",
		Email:          "test@example.com",
		OrganizationID: "test-org",
		PAT:            "test-pat",
	}

	if err := SaveOpenlaneConfig(testConfig); err != nil {
		t.Fatalf("SaveOpenlaneConfig failed: %v", err)
	}

	if !ConfigExists() {
		t.Fatal("Config file should exist after saving")
	}

	if err := DeleteConfig(); err != nil {
		t.Fatalf("DeleteConfig failed: %v", err)
	}

	if ConfigExists() {
		t.Fatal("Config file should not exist after deletion")
	}
}
