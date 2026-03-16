package githubapp

import "testing"

func TestBuilderUsesCanonicalRefs(t *testing.T) {
	t.Parallel()

	definition, err := Builder(Config{})()
	if err != nil {
		t.Fatalf("Builder() error = %v", err)
	}

	if definition.ID != DefinitionID.ID() {
		t.Fatalf("definition ID = %q, want %q", definition.ID, DefinitionID.ID())
	}

	if len(definition.Operations) != 3 {
		t.Fatalf("operations len = %d, want 3", len(definition.Operations))
	}

	if definition.Operations[0].Topic != HealthDefaultOperation.Topic(Slug) {
		t.Fatalf("health topic = %q, want %q", definition.Operations[0].Topic, HealthDefaultOperation.Topic(Slug))
	}

	if definition.Operations[1].Topic != RepositorySyncOperation.Topic(Slug) {
		t.Fatalf("repository topic = %q, want %q", definition.Operations[1].Topic, RepositorySyncOperation.Topic(Slug))
	}

	if definition.Operations[2].Topic != VulnerabilityCollectOperation.Topic(Slug) {
		t.Fatalf("vulnerability topic = %q, want %q", definition.Operations[2].Topic, VulnerabilityCollectOperation.Topic(Slug))
	}
}
