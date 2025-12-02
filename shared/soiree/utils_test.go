package soiree

import (
	"testing"
)

func TestMatchTopicPattern(t *testing.T) {
	tests := []struct {
		pattern string
		subject string
		want    bool
	}{
		// Exact matches
		{"event.some.thing.run", "event.some.thing.run", true},
		// Single node wildcard matches
		{"event.some.*.*", "event.some.thing.run", true},
		{"event.some.*.*", "event.some.thing.meow", true},
		{"event.*", "event.some", true},
		{"event.*", "event.some.thing", false},
		{"event.some.*.run", "event.some.thing.run", true},
		// Single node wildcard non-matches
		{"event.some.*.run", "event.some.thing.meow", false},
		{"event.*.thing.run", "event.some.thing.meow", false},
		{"*.some.thing.run", "event.some.thing.meow", false},
		{"event.some.*.run", "event.some.thing", false},
		// Multi-node wildcard matches
		{"event.some.**", "event.some.thing.run", true},
		{"event.some.**", "event.some.thing.meow", true},
		{"**.thing.run", "event.some.thing.run", true},
		{"event.**", "event.some.thing.run", true},
		{"event.**.run", "event.some.thing.run", true},
		{"**", "event.some.thing.run", true},
		{"**", "event", true},
		{"**", "", true},
		// Multi-node wildcard non-matches
		{"event.**", "event", false},
		{"event.**.run", "event.some.thing.meow", false},
		{"event.some.thing.**", "event.some.other.thing.run", false},
		{"**.thing.run", "event.some.thing.meow", false},
		// Edge cases
		{"*", "", true},
		{"*", "event", true},
		{"*", "event.some", false},
		{"*", "event.some.thing", false},
		{"event.*", "event", false},
		{"event.*", "event.some", true},
		{"event.*", "event.some.thing", false},
		{"**", "", true},
		{"**", "event.some", true},
		{"**", "event.some.thing", true},
		{"**", "event.some.thing.meow", true},
		{"**", "event", true},
		{"event.**", "event.some", true},
		{"event.**", "event.some.thing", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.subject, func(t *testing.T) {
			if got := matchTopicPattern(tt.pattern, tt.subject); got != tt.want {
				t.Errorf("matchTopicPattern(%q, %q) = %v, want %v", tt.pattern, tt.subject, got, tt.want)
			}
		})
	}
}
