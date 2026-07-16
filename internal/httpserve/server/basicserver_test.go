package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	echo "github.com/theopenlane/echox"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()

	r := NewRouter(LogConfig{})
	if r.Echo == nil {
		t.Fatalf("router not properly initialized")
	}

	if r.OAS != nil {
		t.Fatalf("runtime router should carry no OpenAPI state")
	}
}

func TestNewSpecRouter(t *testing.T) {
	t.Parallel()

	r, err := NewSpecRouter(LogConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.Echo == nil || r.OAS == nil || r.SchemaRegistry == nil {
		t.Fatalf("spec router not properly initialized")
	}
}

func TestCustomHTTPErrorHandler(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/hello?x=1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := errors.WithStack(errors.New("boom"))
	CustomHTTPErrorHandler(c, err)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "boom") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}
