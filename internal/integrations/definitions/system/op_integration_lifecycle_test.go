package system

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/theopenlane/core/common/enums"
	generated "github.com/theopenlane/core/internal/ent/generated"
)

func integrationRow(id string, status enums.IntegrationStatus, updatedAt time.Time) *generated.Integration {
	return &generated.Integration{ID: id, Status: status, UpdatedAt: updatedAt}
}

func overrideLifecycleRules(t *testing.T, rules []LifecycleRule) {
	t.Helper()

	original := integrationLifecycleRules
	integrationLifecycleRules = rules

	t.Cleanup(func() { integrationLifecycleRules = original })
}

func TestMatchLifecycleRuleTruthTable(t *testing.T) {
	now := time.Now()

	cases := []struct {
		name      string
		status    enums.IntegrationStatus
		updatedAt time.Time
		wantRule  string
		wantMatch bool
	}{
		{name: "fresh pending", status: enums.IntegrationStatusPending, updatedAt: now, wantMatch: false},
		{name: "in-flight pending", status: enums.IntegrationStatusPending, updatedAt: now.Add(-72 * time.Hour), wantMatch: false},
		{name: "abandoned pending", status: enums.IntegrationStatusPending, updatedAt: now.Add(-(abandonedPendingAge + time.Hour)), wantRule: "reap-abandoned-pending", wantMatch: true},
		{name: "deleted", status: enums.IntegrationStatusDeleted, updatedAt: now, wantRule: "finalize-deleted", wantMatch: true},
		{name: "errored", status: enums.IntegrationStatusErrored, updatedAt: now, wantRule: "reprobe-errored", wantMatch: true},
		{name: "connected", status: enums.IntegrationStatusConnected, updatedAt: now.Add(-2 * abandonedPendingAge), wantMatch: false},
		{name: "disabled", status: enums.IntegrationStatusDisabled, updatedAt: now.Add(-2 * abandonedPendingAge), wantMatch: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rule, matched := matchLifecycleRule(integrationRow("int_1", tc.status, tc.updatedAt), now)

			if matched != tc.wantMatch {
				t.Fatalf("matched = %v, want %v", matched, tc.wantMatch)
			}

			if matched && rule.Name != tc.wantRule {
				t.Fatalf("rule = %q, want %q", rule.Name, tc.wantRule)
			}
		})
	}
}

func TestMatchLifecycleRuleFirstMatchWins(t *testing.T) {
	pendingMatch := func(row *generated.Integration, _ time.Time) bool {
		return row.Status == enums.IntegrationStatusPending
	}

	overrideLifecycleRules(t, []LifecycleRule{
		{Name: "first", Match: pendingMatch, Action: LifecycleActionMark, Reason: "first"},
		{Name: "second", Match: pendingMatch, Action: LifecycleActionRemove},
	})

	rule, matched := matchLifecycleRule(integrationRow("int_1", enums.IntegrationStatusPending, time.Now()), time.Now())

	if !matched || rule.Name != "first" {
		t.Fatalf("rule = %q matched = %v, want first match", rule.Name, matched)
	}
}

func TestDispatchLifecycleActionUnknown(t *testing.T) {
	row := integrationRow("int_unknown", enums.IntegrationStatusConnected, time.Now())

	err := dispatchLifecycleAction(context.Background(), nil, row, LifecycleRule{Name: "bogus", Action: LifecycleAction("bogus")})
	if !errors.Is(err, ErrLifecycleActionUnknown) {
		t.Fatalf("err = %v, want ErrLifecycleActionUnknown", err)
	}
}
