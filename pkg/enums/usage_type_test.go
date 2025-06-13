package enums

import "testing"

func TestToUsageType(t *testing.T) {
	if *ToUsageType("storage") != UsageStorage {
		t.Errorf("expected STORAGE")
	}
	if *ToUsageType("invalid") != UsageInvalid {
		t.Errorf("expected INVALID")
	}
}

func TestUsageTypeValues(t *testing.T) {
	vals := (UsageType("")).Values()
	expected := []string{"STORAGE", "RECORDS", "USERS", "PROGRAMS"}
	if len(vals) != len(expected) {
		t.Fatalf("expected %d values got %d", len(expected), len(vals))
	}
	for i, v := range expected {
		if vals[i] != v {
			t.Errorf("expected %s got %s", v, vals[i])
		}
	}
}
