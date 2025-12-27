package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"path/filepath"
	"testing"
)

func TestComputeSHA(t *testing.T) {
	t.Parallel()
	v := "v0.0.1"
	sum := sha256.Sum256([]byte(v))
	want := hex.EncodeToString(sum[:])
	if got := computeSHA(v); got != want {
		t.Fatalf("computeSHA(%s)=%s, want %s", v, got, want)
	}
}

func TestIsCurrent(t *testing.T) {
	t.Parallel()
	if (*Catalog)(nil).IsCurrent() != true {
		t.Fatal("nil catalog should be current")
	}

	c := New()
	c.Version = "v0.0.1"
	c.SHA = computeSHA(c.Version)
	if !c.IsCurrent() {
		t.Fatal("catalog should be current when SHA matches")
	}
	c.SHA = "mismatch"
	if c.IsCurrent() {
		t.Fatal("catalog should not be current when SHA mismatches")
	}
}

func TestSaveCatalogVersionBump(t *testing.T) {
	t.Parallel()
	c := New()
	c.Version = "v0.0.1"
	c.Modules = FeatureSet{
		"m1": {Billing: Billing{Prices: []Price{{Interval: "month", UnitAmount: 1}}}},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "catalog.yaml")
	if _, err := c.SaveCatalog(path); err != nil {
		t.Fatalf("first save: %v", err)
	}
	if c.Version != "v0.0.2" {
		t.Fatalf("version not bumped: %s", c.Version)
	}
	if c.SHA != computeSHA(c.Version) {
		t.Fatalf("sha not updated")
	}
}
