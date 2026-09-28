package registry

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// credSlot builds a credential slot surface entry with the given schema, backfill declaration, and replaced refs
func credSlot(ref, schema string, backfill bool, replaces ...string) SurfaceCredential {
	return SurfaceCredential{Ref: ref, SurfaceSchema: SurfaceSchema{Schema: json.RawMessage(schema), Backfill: backfill, Replaces: replaces}}
}

// namedRef builds a named surface entry with the given replaced names
func namedRef(name string, replaces ...string) SurfaceNamed {
	return SurfaceNamed{Name: name, Replaces: replaces}
}

// opConfig builds an operation surface entry with the given config schema and replaced names
func opConfig(name, schema string, replaces ...string) SurfaceOperation {
	return SurfaceOperation{Name: name, Schema: json.RawMessage(schema), Replaces: replaces}
}

// findingsText joins findings into one string for substring assertions
func findingsText(findings []string) string {
	return strings.Join(findings, "\n")
}

// surfaceChangeCase is one ClassifySurfaceChange table entry
type surfaceChangeCase struct {
	// name describes the scenario under test
	name string
	// old is the committed surface
	old Surface
	// next is the candidate surface
	next Surface
	// wantErr is the sentinel the returned error must wrap, or nil when no error is expected
	wantErr error
	// wantContains lists substrings every one of which must appear somewhere in the findings
	wantContains []string
	// wantAbsent lists substrings that must not appear anywhere in the findings
	wantAbsent []string
	// wantEmpty requires the findings slice to be empty
	wantEmpty bool
}

// TestClassifySurfaceChange verifies every surface-change rule: destructive removals refuse without a taker, replaced removals and property-level changes are reported as findings, and additive changes are silent or informational
func TestClassifySurfaceChange(t *testing.T) {
	t.Parallel()

	cases := []surfaceChangeCase{
		{
			name:         "credential slot removed with taker converts",
			old:          Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("legacy", `{"type":"object"}`, false)}},
			next:         Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("modern", `{"type":"object"}`, false, "legacy")}},
			wantContains: []string{"stored payloads convert to modern"},
		},
		{
			name:         "credential slot removed without taker refuses",
			old:          Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("legacy", `{"type":"object"}`, false)}},
			next:         Surface{ID: "acme"},
			wantErr:      ErrDestructiveSurfaceChange,
			wantContains: []string{"existing installations cannot resolve it"},
		},
		{
			name:         "operation removed with taker moves history",
			old:          Surface{ID: "acme", Operations: []SurfaceOperation{{Name: "sync.v1"}}},
			next:         Surface{ID: "acme", Operations: []SurfaceOperation{{Name: "sync.v2", Replaces: []string{"sync.v1"}}}},
			wantContains: []string{"run history and health move to sync.v2"},
		},
		{
			name:         "operation removed without taker refuses",
			old:          Surface{ID: "acme", Operations: []SurfaceOperation{{Name: "sync.v1"}}},
			next:         Surface{ID: "acme"},
			wantErr:      ErrDestructiveSurfaceChange,
			wantContains: []string{"its run history and health are orphaned"},
		},
		{
			name:         "webhook removed with taker renames endpoint",
			old:          Surface{ID: "acme", Webhooks: []SurfaceWebhook{{Name: "github"}}},
			next:         Surface{ID: "acme", Webhooks: []SurfaceWebhook{{Name: "github_v2", Replaces: []string{"github"}}}},
			wantContains: []string{"endpoint row renamed to github_v2"},
		},
		{
			name:         "webhook removed without taker refuses",
			old:          Surface{ID: "acme", Webhooks: []SurfaceWebhook{{Name: "github"}}},
			next:         Surface{ID: "acme"},
			wantErr:      ErrDestructiveSurfaceChange,
			wantContains: []string{"its endpoint row is deleted as stale"},
		},
		{
			name: "webhook event removed with taker on a kept webhook",
			old:  Surface{ID: "acme", Webhooks: []SurfaceWebhook{{Name: "github", Events: []SurfaceNamed{namedRef("push")}}}},
			next: Surface{ID: "acme", Webhooks: []SurfaceWebhook{{Name: "github", Events: []SurfaceNamed{namedRef("push_v2", "push")}}}},
			wantContains: []string{
				"replaced by push_v2",
				"allowed events refreshed on upgrade",
			},
		},
		{
			name:         "webhook event removed without taker on a kept webhook is a finding, not a refusal",
			old:          Surface{ID: "acme", Webhooks: []SurfaceWebhook{{Name: "github", Events: []SurfaceNamed{namedRef("push")}}}},
			next:         Surface{ID: "acme", Webhooks: []SurfaceWebhook{{Name: "github"}}},
			wantContains: []string{"allowed events refreshed on upgrade"},
			wantAbsent:   []string{"replaced by"},
		},
		{
			name:         "credential schema property type change is errored",
			old:          Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("cred_a", `{"type":"object","properties":{"foo":{"type":"string"}}}`, false)}},
			next:         Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("cred_a", `{"type":"object","properties":{"foo":{"type":"integer"}}}`, false)}},
			wantContains: []string{"type changed from", outcomeErrored},
		},
		{
			name:         "credential schema enum narrowing is errored",
			old:          Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("cred_a", `{"type":"object","properties":{"level":{"type":"string","enum":["low","medium","high"]}}}`, false)}},
			next:         Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("cred_a", `{"type":"object","properties":{"level":{"type":"string","enum":["low","medium"]}}}`, false)}},
			wantContains: []string{"enum narrowed", outcomeErrored},
		},
		{
			name:         "new required property without default and no backfill is errored",
			old:          Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("cred_a", `{"type":"object","properties":{"name":{"type":"string"}}}`, false)}},
			next:         Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("cred_a", `{"type":"object","properties":{"name":{"type":"string"},"region":{"type":"string"}},"required":["region"]}`, false)}},
			wantContains: []string{"required property region added without default", outcomeErrored},
		},
		{
			name:         "new required property without default but with backfill is a backfill finding",
			old:          Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("cred_a", `{"type":"object","properties":{"name":{"type":"string"}}}`, false)}},
			next:         Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("cred_a", `{"type":"object","properties":{"name":{"type":"string"},"region":{"type":"string"}},"required":["region"]}`, true)}},
			wantContains: []string{"required property region added without default", "backfilled on upgrade"},
			wantAbsent:   []string{outcomeErrored},
		},
		{
			name:      "added optional property produces no finding",
			old:       Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("cred_a", `{"type":"object","properties":{"name":{"type":"string"}}}`, false)}},
			next:      Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("cred_a", `{"type":"object","properties":{"name":{"type":"string"},"nickname":{"type":"string"}}}`, false)}},
			wantEmpty: true,
		},
		{
			name:         "operation config property type change is errored",
			old:          Surface{ID: "acme", Operations: []SurfaceOperation{opConfig("sync.users", `{"type":"object","properties":{"limit":{"type":"integer"}}}`)}},
			next:         Surface{ID: "acme", Operations: []SurfaceOperation{opConfig("sync.users", `{"type":"object","properties":{"limit":{"type":"string"}}}`)}},
			wantContains: []string{"operation sync.users config property limit type changed from", outcomeErrored},
		},
		{
			name:         "new required operation config property without default is errored",
			old:          Surface{ID: "acme", Operations: []SurfaceOperation{opConfig("sync.users", `{"type":"object","properties":{"limit":{"type":"integer"}}}`)}},
			next:         Surface{ID: "acme", Operations: []SurfaceOperation{opConfig("sync.users", `{"type":"object","properties":{"limit":{"type":"integer"},"region":{"type":"string"}},"required":["region"]}`)}},
			wantContains: []string{"operation sync.users config required property region added without default", outcomeErrored},
		},
		{
			name:      "added optional operation config property produces no finding",
			old:       Surface{ID: "acme", Operations: []SurfaceOperation{opConfig("sync.users", `{"type":"object","properties":{"limit":{"type":"integer"}}}`)}},
			next:      Surface{ID: "acme", Operations: []SurfaceOperation{opConfig("sync.users", `{"type":"object","properties":{"limit":{"type":"integer"},"note":{"type":"string"}}}`)}},
			wantEmpty: true,
		},
		{
			name:      "unchanged operation config schema produces no finding",
			old:       Surface{ID: "acme", Operations: []SurfaceOperation{opConfig("sync.users", `{"type":"object","properties":{"limit":{"type":"integer"}}}`)}},
			next:      Surface{ID: "acme", Operations: []SurfaceOperation{opConfig("sync.users", `{"type":"object","properties":{"limit":{"type":"integer"}}}`)}},
			wantEmpty: true,
		},
		{
			name:         "user input added from nil is a change from an empty schema",
			old:          Surface{ID: "acme"},
			next:         Surface{ID: "acme", UserInput: &SurfaceSchema{Schema: json.RawMessage(`{"type":"object","properties":{"region":{"type":"string"}}}`)}},
			wantContains: []string{"user input added: stored config is conformed on upgrade"},
		},
		{
			name:         "installation metadata added is re-derived",
			old:          Surface{ID: "acme"},
			next:         Surface{ID: "acme", Installation: &SurfaceSchema{Schema: json.RawMessage(`{"type":"object"}`)}},
			wantContains: []string{"acme: installation metadata schema added: re-derived on upgrade"},
		},
		{
			name:         "installation metadata removed is re-derived",
			old:          Surface{ID: "acme", Installation: &SurfaceSchema{Schema: json.RawMessage(`{"type":"object"}`)}},
			next:         Surface{ID: "acme"},
			wantContains: []string{"acme: installation metadata schema removed: re-derived on upgrade"},
		},
		{
			name:         "installation metadata changed is re-derived",
			old:          Surface{ID: "acme", Installation: &SurfaceSchema{Schema: json.RawMessage(`{"type":"object"}`)}},
			next:         Surface{ID: "acme", Installation: &SurfaceSchema{Schema: json.RawMessage(`{"type":"string"}`)}},
			wantContains: []string{"acme: installation metadata schema changed: re-derived on upgrade"},
		},
		{
			name:      "unchanged installation metadata produces no finding",
			old:       Surface{ID: "acme", Installation: &SurfaceSchema{Schema: json.RawMessage(`{"type":"object"}`)}},
			next:      Surface{ID: "acme", Installation: &SurfaceSchema{Schema: json.RawMessage(`{"type":"object"}`)}},
			wantEmpty: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			findings, err := ClassifySurfaceChange(tc.old, tc.next)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("ClassifySurfaceChange() error = %v, want wrapping %v", err, tc.wantErr)
				}
			} else if err != nil {
				t.Fatalf("ClassifySurfaceChange() unexpected error = %v", err)
			}

			if tc.wantEmpty && len(findings) != 0 {
				t.Fatalf("ClassifySurfaceChange() findings = %v, want none", findings)
			}

			text := findingsText(findings)

			for _, want := range tc.wantContains {
				if !strings.Contains(text, want) {
					t.Fatalf("ClassifySurfaceChange() findings = %v, want containing %q", findings, want)
				}
			}

			for _, absent := range tc.wantAbsent {
				if strings.Contains(text, absent) {
					t.Fatalf("ClassifySurfaceChange() findings = %v, want not containing %q", findings, absent)
				}
			}
		})
	}
}

// TestGateSurfaceChangeNoExistingSnapshot verifies a committed surface with nothing recorded never refuses, since only removals can be destructive
func TestGateSurfaceChangeNoExistingSnapshot(t *testing.T) {
	t.Parallel()

	next := Surface{ID: "acme", Operations: []SurfaceOperation{{Name: "sync"}}}

	if err := GateSurfaceChange("acme.json", json.RawMessage(`{}`), next); err != nil {
		t.Fatalf("GateSurfaceChange() error = %v, want nil", err)
	}
}

// TestGateSurfaceChangeEqualToNext verifies an unchanged surface produces no findings and no error
func TestGateSurfaceChangeEqualToNext(t *testing.T) {
	t.Parallel()

	next := Surface{
		ID:          "acme",
		Credentials: []SurfaceCredential{credSlot("cred_a", `{"type":"object","properties":{"name":{"type":"string"}}}`, false)},
		Operations:  []SurfaceOperation{{Name: "sync"}},
		Webhooks:    []SurfaceWebhook{{Name: "github", Events: []SurfaceNamed{namedRef("push")}}},
	}

	findings, err := ClassifySurfaceChange(next, next)
	if err != nil {
		t.Fatalf("ClassifySurfaceChange() error = %v, want nil", err)
	}

	if len(findings) != 0 {
		t.Fatalf("ClassifySurfaceChange() findings = %v, want none", findings)
	}

	existing, err := json.Marshal(next)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if err := GateSurfaceChange("acme.json", existing, next); err != nil {
		t.Fatalf("GateSurfaceChange() error = %v, want nil", err)
	}
}

// TestGateSurfaceChangeDestructive verifies a destructive removal against the committed snapshot bytes surfaces ErrDestructiveSurfaceChange
func TestGateSurfaceChangeDestructive(t *testing.T) {
	t.Parallel()

	old := Surface{ID: "acme", Credentials: []SurfaceCredential{credSlot("legacy", `{"type":"object"}`, false)}}

	existing, err := json.Marshal(old)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	next := Surface{ID: "acme"}

	err = GateSurfaceChange("acme.json", existing, next)
	if !errors.Is(err, ErrDestructiveSurfaceChange) {
		t.Fatalf("GateSurfaceChange() error = %v, want wrapping ErrDestructiveSurfaceChange", err)
	}
}
