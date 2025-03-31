package route

import (
	"net/http"
	"net/http/pprof"

	echo "github.com/theopenlane/echox"
)

// registerPPROFroutes registers the pprof routes for the server
func registerPPROFroutes(router *Router) (err error) {
	method := http.MethodGet
	combinedHandlers := map[string]http.HandlerFunc{
		"/debug/pprof":           pprof.Index,
		"/debug/pprof/cmdline":   pprof.Cmdline,
		"/debug/pprof/profile":   pprof.Profile,
		"/debug/pprof/symbol":    pprof.Symbol,
		"/debug/pprof/trace":     pprof.Trace,
		"/debug/pprof/mutex":     pprof.Index,
		"/debug/pprof/allocs":    pprof.Index,
		"/debug/pprof/block":     pprof.Index,
		"/debug/pprof/goroutine": pprof.Index,
		"/debug/pprof/heap":      pprof.Index,
	}

	for path, handler := range combinedHandlers {
		route := echo.Route{
			Name:        path,
			Method:      method,
			Path:        path,
			Middlewares: mw,
			Handler:     echo.WrapHandler(handler),
		}

		if err := router.AddEchoOnlyRoute(route); err != nil {
			return err
		}
	}

	return nil
}
