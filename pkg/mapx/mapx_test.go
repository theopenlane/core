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

func TestAppendOnce(t *testing.T) {
	tests := []struct {
		name string
		ops  []struct{ key, val string }
		want map[string][]string
	}{
		{
			name: "first occurrence is appended",
			ops:  []struct{ key, val string }{{"a", "x"}},
			want: map[string][]string{"a": {"x"}},
		},
		{
			name: "duplicate key is skipped",
			ops:  []struct{ key, val string }{{"a", "x"}, {"a", "y"}},
			want: map[string][]string{"a": {"x"}},
		},
		{
			name: "distinct keys each appended",
			ops:  []struct{ key, val string }{{"a", "x"}, {"b", "y"}},
			want: map[string][]string{"a": {"x"}, "b": {"y"}},
		},
		{
			name: "empty ops produces empty map",
			ops:  []struct{ key, val string }{},
			want: map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := map[string][]string{}
			seen := map[string]struct{}{}

			for _, op := range tt.ops {
				AppendOnce(op.key, op.val, m, seen)
			}

			if !reflect.DeepEqual(tt.want, m) {
				t.Fatalf("expected %v, got %v", tt.want, m)
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

func TestGetOrInit(t *testing.T) {
	type val struct{ n int }

	t.Run("initializes missing key", func(t *testing.T) {
		m := map[string]*val{}
		got := GetOrInit(m, "a", func() *val { return &val{n: 1} })
		if got == nil {
			t.Fatal("expected non-nil value")
		}
		if got.n != 1 {
			t.Fatalf("expected n=1, got n=%d", got.n)
		}
		if m["a"] != got {
			t.Fatal("expected map entry to point to returned value")
		}
	})

	t.Run("returns existing value without calling init", func(t *testing.T) {
		existing := &val{n: 42}
		m := map[string]*val{"a": existing}
		calls := 0
		got := GetOrInit(m, "a", func() *val { calls++; return &val{n: 99} })
		if got != existing {
			t.Fatalf("expected existing value, got %v", got)
		}
		if calls != 0 {
			t.Fatalf("expected init not to be called, got %d calls", calls)
		}
	})

	t.Run("independent keys do not interfere", func(t *testing.T) {
		m := map[string]*val{}
		a := GetOrInit(m, "a", func() *val { return &val{n: 1} })
		b := GetOrInit(m, "b", func() *val { return &val{n: 2} })
		if a == b {
			t.Fatal("expected distinct pointers for different keys")
		}
		if a.n != 1 || b.n != 2 {
			t.Fatalf("expected a.n=1 b.n=2, got a.n=%d b.n=%d", a.n, b.n)
		}
	})
}
