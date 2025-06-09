package serveropts

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestLevelNameHook(t *testing.T) {
	b := &strings.Builder{}
	logger := zerolog.New(b).Hook(LevelNameHook{})
	logger.Log().Msg("test")
	var m map[string]any
	if err := json.Unmarshal([]byte(b.String()), &m); err != nil {
		t.Fatalf("failed to unmarshal log output: %v", err)
	}
	if m[zerolog.LevelFieldName] != zerolog.InfoLevel.String() {
		t.Fatalf("expected level info, got %v", m[zerolog.LevelFieldName])
	}
}
