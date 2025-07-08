package reconciler

import (
	"bytes"
	"context"
	"strings"
	"testing"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/pkg/entitlements"
)

func TestNew_MissingDeps(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
	}{
		{
			name: "missing stripe client",
			opts: []Option{WithDB(&ent.Client{}), WithStripeClient(nil)},
		},
		{
			name: "missing db",
			opts: []Option{WithDB(nil), WithStripeClient(&entitlements.StripeClient{})},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.opts...)
			if err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestAnalyzeOrgNoSubscription(t *testing.T) {
	r := &Reconciler{}
	org := &ent.Organization{}
	action, err := r.analyzeOrg(context.Background(), org)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if action != "create stripe customer & subscription" {
		t.Fatalf("unexpected action: %s", action)
	}
}

func TestPrintRows(t *testing.T) {
	buf := bytes.Buffer{}
	r := &Reconciler{writer: &buf}
	rows := []actionRow{{OrgID: "1", Action: "test"}}
	if err := r.printRows(rows); err != nil {
		t.Fatalf("print rows: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("unexpected line count: %d", len(lines))
	}

	header := strings.Fields(lines[0])
	if len(header) != 2 || header[0] != "ORGANIZATION" || header[1] != "ACTION" {
		t.Fatalf("unexpected header: %q", lines[0])
	}

	row := strings.Fields(lines[1])
	if len(row) != 2 || row[0] != "1" || row[1] != "test" {
		t.Fatalf("unexpected row: %q", lines[1])
	}
}
