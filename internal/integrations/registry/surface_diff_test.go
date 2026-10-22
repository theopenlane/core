package registry

import (
	"errors"
	"testing"
)

// gateCase is one GateSurfaceChange table entry
type gateCase struct {
	// name describes the scenario under test
	name string
	// old is the committed surface
	old Surface
	// next is the candidate surface
	next Surface
	// wantErr is the exact error text expected, empty when the gate passes
	wantErr string
}

// TestGateSurfaceChange verifies the gate passes valid changes and refuses unreplaced removals
func TestGateSurfaceChange(t *testing.T) {
	t.Parallel()

	committed := Surface{
		ID:          "acme",
		Credentials: []SurfaceCredential{{Ref: "cred_a"}},
		Operations:  []SurfaceOperation{{Name: "sync"}},
	}

	cases := []gateCase{
		{
			name: "nothing removed passes",
			old:  committed,
			next: committed,
		},
		{
			name: "additions pass",
			old:  committed,
			next: Surface{
				ID:          "acme",
				Credentials: []SurfaceCredential{{Ref: "cred_a"}, {Ref: "cred_b"}},
				Operations:  []SurfaceOperation{{Name: "sync"}, {Name: "sync.more"}},
			},
		},
		{
			name: "removed slot and operation replaced passes",
			old:  committed,
			next: Surface{
				ID:          "acme",
				Credentials: []SurfaceCredential{{Ref: "cred_v2", Replaces: []string{"cred_a"}}},
				Operations:  []SurfaceOperation{{Name: "sync.v2", Replaces: []string{"sync"}}},
			},
		},
		{
			name:    "unreplaced credential slot refuses",
			old:     committed,
			next:    Surface{ID: "acme", Operations: committed.Operations},
			wantErr: ErrDestructiveSurfaceChange.Error() + ": acme: credential cred_a",
		},
		{
			name:    "unreplaced operation refuses",
			old:     committed,
			next:    Surface{ID: "acme", Credentials: committed.Credentials},
			wantErr: ErrDestructiveSurfaceChange.Error() + ": acme: operation sync",
		},
		{
			name: "every unreplaced name across kinds is listed in one error",
			old: Surface{
				ID:          "acme",
				Credentials: []SurfaceCredential{{Ref: "cred_a"}, {Ref: "cred_b"}, {Ref: "cred_c"}},
				Operations:  []SurfaceOperation{{Name: "sync"}, {Name: "sync.more"}},
			},
			next: Surface{
				ID:          "acme",
				Credentials: []SurfaceCredential{{Ref: "cred_d", Replaces: []string{"cred_c"}}},
			},
			wantErr: ErrDestructiveSurfaceChange.Error() + ": acme: credential cred_a, cred_b; operation sync, sync.more",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := GateSurfaceChange(tc.old, tc.next)

			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("GateSurfaceChange() error = %v, want nil", err)
			case tc.wantErr == "":
				return
			case !errors.Is(err, ErrDestructiveSurfaceChange):
				t.Fatalf("GateSurfaceChange() error = %v, want wrapping ErrDestructiveSurfaceChange", err)
			case err.Error() != tc.wantErr:
				t.Fatalf("GateSurfaceChange() error = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

// TestGateSurfaceChangeEmptySnapshot verifies a committed snapshot recording nothing never refuses
func TestGateSurfaceChangeEmptySnapshot(t *testing.T) {
	t.Parallel()

	if err := GateSurfaceChange(Surface{}, Surface{ID: "acme", Operations: []SurfaceOperation{{Name: "sync"}}}); err != nil {
		t.Fatalf("GateSurfaceChange() error = %v, want nil", err)
	}
}
