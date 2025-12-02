package logx_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/shared/logx"
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
