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
	r, err := NewRouter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Echo == nil || r.OAS == nil {
		t.Fatalf("router not properly initialized")
	}
}

func TestCustomHTTPErrorHandler(t *testing.T) {
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
