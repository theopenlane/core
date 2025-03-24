package echolog_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/pkg/logx/echolog"
)

func TestCtx(t *testing.T) {
	b := &bytes.Buffer{}
	l := echolog.New(b)
	zerologger := l.Unwrap()
	ctx := l.WithContext(context.Background())

	assert.Equal(t, echolog.Ctx(ctx), &zerologger)
}
