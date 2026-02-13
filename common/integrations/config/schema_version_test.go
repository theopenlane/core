package config

import "testing"

func TestSupportsSchemaVersion(t *testing.T) {
	var spec *ProviderSpec
	if spec.supportsSchemaVersion() {
		t.Fatalf("expected nil spec to be unsupported")
	}

	spec = &ProviderSpec{}
	if !spec.supportsSchemaVersion() {
		t.Fatalf("expected default version to be supported")
	}

	spec.SchemaVersion = "v9"
	if spec.supportsSchemaVersion() {
		t.Fatalf("expected unsupported version")
	}
}
