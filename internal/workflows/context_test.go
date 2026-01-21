package workflows

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithContext(t *testing.T) {
	ctx := context.Background()

	assert.False(t, IsWorkflowBypass(ctx))

	ctx = WithContext(ctx)

	assert.True(t, IsWorkflowBypass(ctx))
}

func TestFromContext(t *testing.T) {
	testCases := []struct {
		name     string
		ctx      context.Context
		expected bool
	}{
		{
			name:     "context without workflow bypass",
			ctx:      context.Background(),
			expected: false,
		},
		{
			name:     "context with workflow bypass",
			ctx:      WithContext(context.Background()),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, ok := FromContext(tc.ctx)
			assert.Equal(t, tc.expected, ok)
		})
	}
}

func TestIsWorkflowBypass(t *testing.T) {
	testCases := []struct {
		name     string
		ctx      context.Context
		expected bool
	}{
		{
			name:     "no bypass in context",
			ctx:      context.Background(),
			expected: false,
		},
		{
			name:     "bypass in context",
			ctx:      WithContext(context.Background()),
			expected: true,
		},
		{
			name:     "bypass persists through context chain",
			ctx:      context.WithValue(WithContext(context.Background()), "key", "value"),
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsWorkflowBypass(tc.ctx)
			assert.Equal(t, tc.expected, result)
		})
	}
}
