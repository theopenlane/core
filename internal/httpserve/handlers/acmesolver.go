package handlers

import (
	"net/http"

	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/dnsverification"
	"github.com/theopenlane/ent/generated/privacy"
	"github.com/theopenlane/iam/auth"
	models "github.com/theopenlane/shared/openapi"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/rout"
)

// ACMESolverHandler handles ACME challenge requests by looking up the challenge path
// and returning the expected challenge value for domain verification
func (h *Handler) ACMESolverHandler(ctx echo.Context, openapi *OpenAPIContext) error {
	in, err := BindAndValidateWithAutoRegistry(ctx, h, openapi.Operation, models.ExampleAcmeSolverRequest, rout.Reply{}, openapi.Registry)
	if err != nil {
		return h.InvalidInput(ctx, err, openapi)
	}

	if isRegistrationContext(ctx) {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx.Request().Context(), privacy.Allow) // bypass privacy policy
	allowCtx = contextx.With(allowCtx, auth.AcmeSolverContextKey{})

	res, err := h.DBClient.DNSVerification.Query().Where(
		dnsverification.AcmeChallengePathEQ(in.Path),
		dnsverification.DeletedAtIsNil(),
	).First(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.NotFound(ctx, err)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest, openapi)
	}

	return ctx.String(http.StatusOK, res.ExpectedAcmeChallengeValue)
}
