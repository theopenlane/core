package operations

import (
	"testing"

	"github.com/theopenlane/core/common/integrations/types"
)

func TestPaginationEffectivePageSize(t *testing.T) {
	p := Pagination{PerPage: 25, PageSize: 10}
	if got := p.EffectivePageSize(5); got != 25 {
		t.Fatalf("expected per_page to win, got %d", got)
	}
	p = Pagination{PageSize: 10}
	if got := p.EffectivePageSize(5); got != 10 {
		t.Fatalf("expected page_size to win, got %d", got)
	}
	p = Pagination{}
	if got := p.EffectivePageSize(5); got != 5 {
		t.Fatalf("expected default to win, got %d", got)
	}
}

func TestRepositorySelectorList(t *testing.T) {
	selector := RepositorySelector{
		Repositories: []types.TrimmedString{"alpha", "beta"},
		Repos:        []types.TrimmedString{"beta", "gamma"},
		Repository:   types.TrimmedString("delta"),
	}
	got := selector.List()
	want := []string{"alpha", "beta", "gamma", "delta"}
	if len(got) != len(want) {
		t.Fatalf("expected %d repos, got %d", len(want), len(got))
	}
	for i, value := range want {
		if got[i] != value {
			t.Fatalf("expected %q at %d, got %q", value, i, got[i])
		}
	}
}
