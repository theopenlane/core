package logx_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/logx"
)

func TestCtx(t *testing.T) {
	b := &bytes.Buffer{}
	l := logx.Configure(logx.LoggerConfig{
		Writer:   b,
		WithEcho: true,
	}).Echo
	zerologger := l.Unwrap()
	ctx := l.WithContext(context.Background())

	assert.Equal(t, logx.Ctx(ctx), &zerologger)
}

func TestWithField(t *testing.T) {
	ctx := context.Background()

	ctx = logx.WithField(ctx, "request_id", "req_123")

	fields := logx.FieldsFromContext(ctx)
	assert.NotNil(t, fields)
	assert.Equal(t, "req_123", fields["request_id"])
}

func TestWithFields(t *testing.T) {
	ctx := context.Background()

	ctx = logx.WithFields(ctx, map[string]any{
		"request_id": "req_456",
		"user_id":    "user_789",
	})

	fields := logx.FieldsFromContext(ctx)
	assert.NotNil(t, fields)
	assert.Equal(t, "req_456", fields["request_id"])
	assert.Equal(t, "user_789", fields["user_id"])
}

func TestFieldsFromContextEmpty(t *testing.T) {
	ctx := context.Background()

	fields := logx.FieldsFromContext(ctx)
	assert.Nil(t, fields)
}

func TestFieldsFromContextNil(t *testing.T) {
	fields := logx.FieldsFromContext(nil)
	assert.Nil(t, fields)
}

func TestWithFieldsEmpty(t *testing.T) {
	ctx := context.Background()
	original := ctx

	ctx = logx.WithFields(ctx, nil)
	assert.Equal(t, original, ctx)

	ctx = logx.WithFields(ctx, map[string]any{})
	assert.Equal(t, original, ctx)
}

func TestWithFieldAccumulates(t *testing.T) {
	ctx := context.Background()

	ctx = logx.WithField(ctx, "field1", "value1")
	ctx = logx.WithField(ctx, "field2", "value2")

	fields := logx.FieldsFromContext(ctx)
	assert.NotNil(t, fields)
	assert.Equal(t, "value1", fields["field1"])
	assert.Equal(t, "value2", fields["field2"])
}
