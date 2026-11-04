package models

import (
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func TestSSOAuthorizationMap_MarshalGQL(t *testing.T) {
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	m := SSOAuthorizationMap{"org": ts}
	var sb strings.Builder

	m.MarshalGQL(&sb)

	want := `{"org":"2024-01-01T00:00:00Z"}`
	assert.Check(t, is.Equal(want, sb.String()))
}

func TestSSOAuthorizationMap_UnmarshalGQL(t *testing.T) {
	in := map[string]any{"org": "2024-01-01T00:00:00Z"}
	var m SSOAuthorizationMap
	err := m.UnmarshalGQL(in)
	assert.NilError(t, err)

	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Check(t, is.DeepEqual(SSOAuthorizationMap{"org": ts}, m))
}
