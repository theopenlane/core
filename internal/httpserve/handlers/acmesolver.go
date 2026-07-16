package handlers

import (
	"net/http"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/dnsverification"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
)

// ACMESolverHandler handles ACME challenge requests by looking up the challenge path
// and returning the expected challenge value for domain verification
func (h *Handler) ACMESolverHandler(ctx echo.Context) error {
	in, err := BindAndValidate[models.AcmeSolverRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	allowCtx := privacy.DecisionContext(ctx.Request().Context(), privacy.Allow) // bypass privacy policy
	allowCtx = auth.WithCaller(allowCtx, auth.NewAcmeSolverCaller(""))

	res, err := h.DBClient.DNSVerification.Query().Where(
		dnsverification.AcmeChallengePathEQ(in.Path),
		dnsverification.DeletedAtIsNil(),
	).First(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return h.NotFound(ctx, err)
		}

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	return ctx.String(http.StatusOK, res.ExpectedAcmeChallengeValue)
}
