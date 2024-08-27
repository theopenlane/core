package route

import (
	"net/http"

	echo "github.com/theopenlane/echox"
)

// registerEventPublisher registers the event publisher endpoint
func registerEventPublisher(router *Router) (err error) {
	path := "/event/publish"
	method := http.MethodPost
	name := "EventPublisher"

	route := echo.Route{
		Name:        name,
		Method:      method,
		Path:        path,
		Middlewares: authMW,
		Handler: func(c echo.Context) error {
			return router.Handler.EventPublisher(c)
		},
	}

	eventOperation := router.Handler.BindEventPublisher()

	if err := router.Addv1Route(path, method, eventOperation, route); err != nil {
		return err
	}

	return nil
}
