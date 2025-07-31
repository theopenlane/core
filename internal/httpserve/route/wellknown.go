package route

import (
	"embed"
	"net/http"

	echo "github.com/theopenlane/echox"

	"github.com/theopenlane/core/internal/httpserve/handlers"
)

//go:embed webauthn
var webauthn embed.FS

// registerWebAuthnWellKnownHandler serves up the webauthn well-known file
func registerWebAuthnWellKnownHandler(router *Router) (err error) {
	config := Config{
		Path:        "/.well-known/webauthn",
		Method:      http.MethodGet,
		Name:        "WebAuthnWellKnown",
		Description: "WebAuthn well-known configuration file",
		Tags:        []string{"well-known", "webauthn"},
		OperationID: "WebAuthnWellKnown",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			return echo.StaticFileHandler("webauthn", webauthn)(ctx)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}

// registerJwksWellKnownHandler supplies the JWKS endpoint
func registerJwksWellKnownHandler(router *Router) (err error) {
	config := Config{
		Path:        "/.well-known/jwks.json",
		Method:      http.MethodGet,
		Name:        "JWKS",
		Description: "JSON Web Key Set for JWT token validation",
		Tags:        []string{"well-known", "authentication"},
		OperationID: "JWKS",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			return ctx.JSON(http.StatusOK, router.Handler.JWTKeys)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}

//go:embed security.txt
var securityTxt embed.FS

// registerSecurityTxtHandler serves up the text output of the security.txt
func registerSecurityTxtHandler(router *Router) (err error) {
	config := Config{
		Path:        "/.well-known/security.txt",
		Method:      http.MethodGet,
		Name:        "SecurityTxt",
		Description: "Security contact information and vulnerability disclosure policy",
		Tags:        []string{"well-known", "security"},
		OperationID: "SecurityTxt",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			return echo.StaticFileHandler("security.txt", securityTxt)(ctx)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}

//go:embed applemerchant
var applemerchant embed.FS

// registerAppleMerchantHandler serves up the text output of the applemerchant file
func registerAppleMerchantHandler(router *Router) (err error) {
	config := Config{
		Path:        "/.well-known/apple-developer-merchantid-domain-association",
		Method:      http.MethodGet,
		Name:        "AppleMerchant",
		Description: "Apple Developer Merchant ID domain association file",
		Tags:        []string{"well-known", "payments"},
		OperationID: "AppleMerchant",
		Security:    handlers.PublicSecurity,
		Middlewares: *publicEndpoint,
		SimpleHandler: func(ctx echo.Context) error {
			return echo.StaticFileHandler("applemerchant", applemerchant)(ctx)
		},
	}

	return router.AddUnversionedHandlerRoute(config)
}
