package operations

import (
	"testing"
	"time"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
)

func TestNewestDirectoryAccount(t *testing.T) {
	t.Parallel()

	older := &ent.DirectoryAccount{ID: "older", CreatedAt: time.Now().Add(-time.Hour)}
	newer := &ent.DirectoryAccount{ID: "newer", CreatedAt: time.Now()}

	if got := newestDirectoryAccount([]*ent.DirectoryAccount{older, newer}); got == nil || got.ID != "newer" {
		t.Fatalf("got %+v, want the most recently created candidate", got)
	}

	if got := newestDirectoryAccount(nil); got != nil {
		t.Fatalf("got %+v, want nil for no candidates", got)
	}
}
