package mapx

import (
	"reflect"
	"testing"
)

func TestDeepCloneMapAny(t *testing.T) {
	if got := DeepCloneMapAny(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}

	src := map[string]any{
		"string": "value",
		"nested": map[string]any{"key": "value"},
		"list": []any{
			map[string]any{"id": 1},
			"a",
		},
		"labels": map[string]string{"env": "prod"},
		"tags":   []string{"tag1", "tag2"},
	}

	clone := DeepCloneMapAny(src)
	if !reflect.DeepEqual(src, clone) {
		t.Fatalf("expected %v, got %v", src, clone)
	}

	clone["string"] = "changed"
	clone["nested"].(map[string]any)["key"] = "changed"
	clone["list"].([]any)[0].(map[string]any)["id"] = 999
	clone["labels"].(map[string]string)["env"] = "dev"
	clone["tags"].([]string)[0] = "changed"

	if src["string"] == "changed" {
		t.Fatal("expected top-level value isolation")
	}
	if src["nested"].(map[string]any)["key"] == "changed" {
		t.Fatal("expected nested map isolation")
	}
	if src["list"].([]any)[0].(map[string]any)["id"] == 999 {
		t.Fatal("expected nested slice+map isolation")
	}
	if src["labels"].(map[string]string)["env"] == "dev" {
		t.Fatal("expected map[string]string isolation")
	}
	if src["tags"].([]string)[0] == "changed" {
		t.Fatal("expected []string isolation")
	}
}

func TestCloneMapStringSlice(t *testing.T) {
	if got := CloneMapStringSlice(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := CloneMapStringSlice(map[string][]string{}); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}

	src := map[string][]string{"edgeA": {"id1", "id2"}}
	clone := CloneMapStringSlice(src)
	if !reflect.DeepEqual(src, clone) {
		t.Fatalf("expected %v, got %v", src, clone)
	}

	clone["edgeA"][0] = "changed"
	if src["edgeA"][0] == "changed" {
		t.Fatal("expected slice copy isolation")
	}

	if got := CloneMapStringSlice(map[string][]string{" ": {"id1"}}); got != nil {
		t.Fatalf("expected nil for blank keys, got %v", got)
	}
}

func TestPruneMapZeroAny(t *testing.T) {
	input := map[string]any{
		"displayName": "GitHub Enterprise",
		"emptyString": "",
		"zeroInt":     0,
		"nilValue":    nil,
		"boolFalse":   false,
		"nested": map[string]any{
			"drop": "",
			"keep": "value",
			"deep": map[string]any{
				"dropToo": 0,
			},
		},
	}

	got := PruneMapZeroAny(input)
	expected := map[string]any{
		"displayName": "GitHub Enterprise",
		"boolFalse":   false,
		"nested": map[string]any{
			"keep": "value",
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestDeepMergeMapAny(t *testing.T) {
	base := map[string]any{
		"displayName": "GitHub",
		"oauth": map[string]any{
			"clientId": "base-id",
			"scopes":   []string{"repo"},
			"params": map[string]any{
				"allow_signup": "true",
			},
		},
	}
	override := map[string]any{
		"oauth": map[string]any{
			"clientId": "override-id",
			"params": map[string]any{
				"prompt": "consent",
			},
		},
	}

	got := DeepMergeMapAny(base, override)
	expected := map[string]any{
		"displayName": "GitHub",
		"oauth": map[string]any{
			"clientId": "override-id",
			"scopes":   []string{"repo"},
			"params": map[string]any{
				"allow_signup": "true",
				"prompt":       "consent",
			},
		},
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestMapSetFromSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  map[string]struct{}
	}{
		{name: "empty", input: []string{}, want: map[string]struct{}{}},
		{name: "dedup", input: []string{"foo", "bar", "foo"}, want: map[string]struct{}{"foo": {}, "bar": {}}},
		{name: "empty string", input: []string{""}, want: map[string]struct{}{"": {}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapSetFromSlice(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestMapIntersectionUnique(t *testing.T) {
	tests := []struct {
		name  string
		left  []string
		right []string
		want  []string
	}{
		{name: "both empty", left: []string{}, right: []string{}, want: []string{}},
		{name: "no overlap", left: []string{"foo"}, right: []string{"bar"}, want: []string{}},
		{name: "preserves right order", left: []string{"foo", "bar", "baz"}, right: []string{"baz", "foo"}, want: []string{"baz", "foo"}},
		{name: "dedups right", left: []string{"foo", "bar"}, right: []string{"foo", "foo", "bar"}, want: []string{"foo", "bar"}},
		{name: "empty string", left: []string{""}, right: []string{""}, want: []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapIntersectionUnique(tt.left, tt.right)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
