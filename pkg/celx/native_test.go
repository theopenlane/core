package celx

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nativeTestEntity is a flat projection-shaped struct: read-only id plus scalar/FK fields,
// snake_case json tags — the shape the generated entity projection takes.
type nativeTestEntity struct {
	ID               string `json:"id"`
	IdentityHolderID string `json:"identity_holder_id"`
	Category         string `json:"category"`
	PrimaryDirectory bool   `json:"primary_directory"`
}

func nativeEval(t *testing.T, withSource bool) *NativeEntityEvaluator {
	t.Helper()

	var sourceType reflect.Type
	if withSource {
		sourceType = reflect.TypeFor[nativeTestEntity]()
	}

	ev, err := NewNativeEntityEvaluator(StrictEnvConfig(), FastEvalConfig(), reflect.TypeFor[nativeTestEntity](), sourceType)
	require.NoError(t, err)

	return ev
}

// The exact real cross-entity expression, which breaks against Create/UpdateXInput (no id field)
// but must resolve against the projection type.
func TestNativeEntity_SourceID(t *testing.T) {
	ev := nativeEval(t, true)

	target, _ := json.Marshal(nativeTestEntity{IdentityHolderID: "ih-1"})
	source, _ := json.Marshal(nativeTestEntity{ID: "ih-1"})

	match, err := ev.EvaluateBoolWithSource(context.Background(), "target.identity_holder_id == source.id", target, source)
	require.NoError(t, err)
	assert.True(t, match)
}

func TestNativeEntity_SourceID_NoMatch(t *testing.T) {
	ev := nativeEval(t, true)

	target, _ := json.Marshal(nativeTestEntity{IdentityHolderID: "ih-1"})
	source, _ := json.Marshal(nativeTestEntity{ID: "ih-2"})

	match, err := ev.EvaluateBoolWithSource(context.Background(), "target.identity_holder_id == source.id", target, source)
	require.NoError(t, err)
	assert.False(t, match)
}

func TestNativeEntity_TargetOnly(t *testing.T) {
	ev := nativeEval(t, false)

	target, _ := json.Marshal(nativeTestEntity{ID: "x", PrimaryDirectory: true})

	match, err := ev.EvaluateBool(context.Background(), `target.primary_directory && target.id == "x"`, target)
	require.NoError(t, err)
	assert.True(t, match)
}

// A misspelled field fails at compile, unlike the DynType+map path which accepts it silently.
func TestNativeEntity_UnknownFieldRejected(t *testing.T) {
	ev := nativeEval(t, false)

	target, _ := json.Marshal(nativeTestEntity{ID: "x"})

	_, err := ev.EvaluateBool(context.Background(), `target.identifier == "x"`, target)
	require.Error(t, err)
}
