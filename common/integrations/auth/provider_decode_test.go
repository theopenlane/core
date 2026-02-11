package auth

import "testing"

type decodeTarget struct {
	RoleARN string `json:"roleArn"`
	Count   int    `json:"count"`
}

func TestDecodeProviderData_NormalizesKeys(t *testing.T) {
	config := map[string]any{
		"role_arn": "arn:test",
		"COUNT":    "5",
		"extra":    "ignored",
	}

	var target decodeTarget
	if err := DecodeProviderData(config, &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.RoleARN != "arn:test" {
		t.Fatalf("expected role arn, got %q", target.RoleARN)
	}
	if target.Count != 5 {
		t.Fatalf("expected count 5, got %d", target.Count)
	}
}

func TestDecodeProviderData_TargetNil(t *testing.T) {
	if err := DecodeProviderData(map[string]any{"foo": "bar"}, nil); err != ErrDecodeProviderDataTargetNil {
		t.Fatalf("expected ErrDecodeProviderDataTargetNil, got %v", err)
	}
}
