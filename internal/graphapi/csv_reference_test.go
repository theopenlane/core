package graphapi

import (
	"context"
	"sort"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

// csvRefRow is a minimal input struct for csv reference mapping tests.
type csvRefRow struct {
	UserID     string
	UserIDs    []string
	UserEmail  string
	UserEmails []string
}

// TestResolveCSVReferenceRulesSuccess verifies successful reference resolution.
func TestResolveCSVReferenceRulesSuccess(t *testing.T) {
	t.Parallel()

	rows := []*csvRefRow{
		{
			UserEmail:  "Test@Example.com",
			UserEmails: []string{"a@example.com", "b@example.com"},
			UserIDs:    []string{"existing"},
		},
	}

	lookup := func(_ context.Context, _ []string) (map[string]string, error) {
		return map[string]string{
			"test@example.com": "id-1",
			"a@example.com":    "id-2",
			"b@example.com":    "id-3",
		}, nil
	}

	rules := []CSVReferenceRule{
		{SourceField: "UserEmail", TargetField: "UserID", Lookup: lookup},
		{SourceField: "UserEmails", TargetField: "UserIDs", Lookup: lookup},
	}

	err := resolveCSVReferenceRules(context.Background(), rows, rules...)
	assert.NilError(t, err)

	assert.Check(t, cmp.Equal(rows[0].UserID, "id-1"))
	expected := []string{"existing", "id-2", "id-3"}
	actual := append([]string(nil), rows[0].UserIDs...)
	sort.Strings(actual)
	sort.Strings(expected)
	assert.Check(t, cmp.DeepEqual(actual, expected))
}

// TestResolveCSVReferenceRulesMissing ensures missing references return validation errors.
func TestResolveCSVReferenceRulesMissing(t *testing.T) {
	t.Parallel()

	rows := []*csvRefRow{{UserEmail: "missing@example.com"}}
	rules := []CSVReferenceRule{
		{
			SourceField: "UserEmail",
			TargetField: "UserID",
			Lookup: func(_ context.Context, _ []string) (map[string]string, error) {
				return map[string]string{}, nil
			},
		},
	}

	err := resolveCSVReferenceRules(context.Background(), rows, rules...)
	assert.ErrorContains(t, err, "unable to resolve userEmail")
}

// TestResolveCSVReferenceRulesCreate verifies create callbacks are applied.
func TestResolveCSVReferenceRulesCreate(t *testing.T) {
	t.Parallel()

	rows := []*csvRefRow{{UserEmail: "created@example.com"}}
	rule := CSVReferenceRule{
		SourceField: "UserEmail",
		TargetField: "UserID",
		Lookup: func(_ context.Context, _ []string) (map[string]string, error) {
			return map[string]string{}, nil
		},
		Create: func(_ context.Context, values []string) (map[string]string, error) {
			resolved := map[string]string{}
			for _, value := range values {
				key := normalizeCSVReferenceKey(value)
				resolved[key] = "created-id"
			}
			return resolved, nil
		},
	}

	err := resolveCSVReferenceRules(context.Background(), rows, rule)
	assert.NilError(t, err)
	assert.Check(t, cmp.Equal(rows[0].UserID, "created-id"))
}

// TestResolveCSVReferencesForSchemaWithoutRules verifies no error for schemas without rules.
func TestResolveCSVReferencesForSchemaWithoutRules(t *testing.T) {
	t.Parallel()

	type testRow struct{ Name string }
	rows := []*testRow{{Name: "test"}}
	err := resolveCSVReferencesForSchema(context.Background(), "User", rows)
	assert.NilError(t, err)
}

// TestResolveCSVReferencesForSchemaNonexistent verifies no error for nonexistent schemas.
func TestResolveCSVReferencesForSchemaNonexistent(t *testing.T) {
	t.Parallel()

	type testRow struct{ Name string }
	rows := []*testRow{{Name: "test"}}
	err := resolveCSVReferencesForSchema(context.Background(), "NonexistentSchema", rows)
	assert.NilError(t, err)
}
