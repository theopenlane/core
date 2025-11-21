package schema

import (
	"os"
	"reflect"
	"sort"
	"testing"
)

// defaultCreateObjectTypes is kept for test reference only (not used in production logic)
var defaultCreateObjectTypes = []string{
	"control",
	"control_implementation",
	"control_objective",
	"evidence",
	"group",
	"internal_policy",
	"mapped_control",
	"narrative",
	"procedure",
	"program",
	"risk",
	"scheduled_job",
	"standard",
	"template",
	"subprocessor",
}

// TestCreatorTypesFromModel_ReferenceDefault is a reference test to compare the parsed model output
// to the legacy defaultCreateObjectTypes list. This is for informational purposes only and not for production logic.
func TestCreatorTypesFromModel_ReferenceDefault(t *testing.T) {
	model, err := os.ReadFile(fgaModelPath)
	if err != nil {
		t.Skipf("skipping: failed to read FGA model: %v", err)
	}

	parsed, err := creatorTypesFromModel(model)
	if err != nil {
		t.Fatalf("failed to parse creator types from model: %v", err)
	}

	parsedSet := make(map[string]struct{}, len(parsed))
	for _, v := range parsed {
		parsedSet[v] = struct{}{}
	}
	defaultSet := make(map[string]struct{}, len(defaultCreateObjectTypes))
	for _, v := range defaultCreateObjectTypes {
		defaultSet[v] = struct{}{}
	}

	missingInParsed := []string{}
	missingInDefault := []string{}
	for v := range defaultSet {
		if _, ok := parsedSet[v]; !ok {
			missingInParsed = append(missingInParsed, v)
		}
	}
	for v := range parsedSet {
		if _, ok := defaultSet[v]; !ok {
			missingInDefault = append(missingInDefault, v)
		}
	}

	sort.Strings(missingInParsed)
	sort.Strings(missingInDefault)

	if len(missingInParsed) > 0 || len(missingInDefault) > 0 {
		t.Logf("Difference between defaultCreateObjectTypes and parsed FGA model\n")
		if len(missingInParsed) > 0 {
			t.Logf("In defaultCreateObjectTypes but not in parsed: %v", missingInParsed)
		}
		if len(missingInDefault) > 0 {
			t.Logf("In parsed but not in defaultCreateObjectTypes: %v", missingInDefault)
		}
	}
}

func TestCreatorTypesFromModel_AddType(t *testing.T) {
	// Simulate a model with an extra creator type
	model := `# FGA Model
type organization
define control_creator: group#member
define evidence_creator: group#member
define newtype_creator: group#member
`
	parsed, err := creatorTypesFromModel([]byte(model))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sort.Strings(parsed)
	expected := []string{"control", "evidence", "newtype"}
	sort.Strings(expected)
	if !reflect.DeepEqual(parsed, expected) {
		t.Errorf("expected %v, got %v", expected, parsed)
	}
}

func TestCreatorTypesFromModel_RemoveType(t *testing.T) {
	// Simulate a model with one type removed
	model := `# FGA Model
type organization
define control_creator: group#member
# define evidence_creator: group#member
`
	parsed, err := creatorTypesFromModel([]byte(model))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []string{"control"}
	if !reflect.DeepEqual(parsed, expected) {
		t.Errorf("expected %v, got %v", expected, parsed)
	}
}
