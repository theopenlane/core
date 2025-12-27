package models

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSOAuthorizationMap_MarshalGQL(t *testing.T) {
	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	m := SSOAuthorizationMap{"org": ts}
	var sb strings.Builder

	m.MarshalGQL(&sb)

	want := `{"org":"2024-01-01T00:00:00Z"}`
	assert.Equal(t, want, sb.String())
}

func TestSSOAuthorizationMap_UnmarshalGQL(t *testing.T) {
	in := map[string]any{"org": "2024-01-01T00:00:00Z"}
	var m SSOAuthorizationMap
	err := m.UnmarshalGQL(in)
	require.NoError(t, err)

	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, SSOAuthorizationMap{"org": ts}, m)
}
