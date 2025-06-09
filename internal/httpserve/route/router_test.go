package route

import (
	"net/http"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	echo "github.com/theopenlane/echox"
)

func newTestRouter() *Router {
	return &Router{
		Echo: echo.New(),
		OAS:  &openapi3.T{Paths: openapi3.NewPaths()},
	}
}

func TestAddRoute(t *testing.T) {
	r := newTestRouter()
	op := &openapi3.Operation{OperationID: "op"}
	rt := echo.Route{Path: "/t", Method: http.MethodGet, Handler: func(echo.Context) error { return nil }}
	if err := r.AddRoute("/t", http.MethodGet, op, rt); err != nil {
		t.Fatalf("add route failed: %v", err)
	}
	if r.OAS.Paths.Find("/t") == nil {
		t.Fatalf("path not registered")
	}
}

func TestAddV1Route(t *testing.T) {
	r := newTestRouter()
	op := &openapi3.Operation{OperationID: "v1"}
	rt := echo.Route{Path: "/v1/t", Method: http.MethodGet, Handler: func(echo.Context) error { return nil }}
	if err := r.AddV1Route("/v1/t", http.MethodGet, op, rt); err != nil {
		t.Fatalf("add v1 route failed: %v", err)
	}
	if r.OAS.Paths.Find("/v1/t") == nil {
		t.Fatalf("path not registered")
	}
}

func TestAddUnversionedRoute(t *testing.T) {
	r := newTestRouter()
	op := &openapi3.Operation{OperationID: "u"}
	rt := echo.Route{Path: "/u", Method: http.MethodGet, Handler: func(echo.Context) error { return nil }}
	if err := r.AddUnversionedRoute("/u", http.MethodGet, op, rt); err != nil {
		t.Fatalf("add unversioned route failed: %v", err)
	}
	if r.OAS.Paths.Find("/u") == nil {
		t.Fatalf("path not registered")
	}
}

func TestAddEchoOnlyRoute(t *testing.T) {
	r := newTestRouter()
	rt := echo.Route{Path: "/e", Method: http.MethodGet, Handler: func(echo.Context) error { return nil }}
	if err := r.AddEchoOnlyRoute(rt); err != nil {
		t.Fatalf("add echo route failed: %v", err)
	}
	if r.OAS.Paths.Find("/e") != nil {
		t.Fatalf("should not add to spec")
	}
}
