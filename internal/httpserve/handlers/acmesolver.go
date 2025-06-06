package handlers

import (
	"net/http"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/dnsverification"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/models"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
)

// ACMESolverHandler handles ACME challenge requests by looking up the challenge path
// and returning the expected challenge value for domain verification
func (h *Handler) ACMESolverHandler(ctx echo.Context) error {
	var in models.AcmeSolverRequest
	if err := ctx.Bind(&in); err != nil {
		return h.InvalidInput(ctx, err)
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

		return h.InternalServerError(ctx, err)
	}

	return ctx.String(http.StatusOK, res.ExpectedAcmeChallengeValue)
}
