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
	l := logx.New(b)
	zerologger := l.Unwrap()
	ctx := l.WithContext(context.Background())

	assert.Equal(t, logx.Ctx(ctx), &zerologger)
}
