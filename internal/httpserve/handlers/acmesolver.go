package handlers

import (
	"net/http"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/dnsverification"
	"github.com/theopenlane/core/v2/internal/ent/privacy/rule"
	echo "github.com/theopenlane/echox"
)

// ACMESolverHandler handles ACME challenge requests by looking up the challenge path
// and returning the expected challenge value for domain verification
func (h *Handler) ACMESolverHandler(ctx echo.Context) error {
	in, err := BindAndValidate[models.AcmeSolverRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	// unauthenticated lookup by challenge path before the owning org is known, so it only needs a cross-org internal read
	allowCtx := rule.WithInternalCrossOrgContext(ctx.Request().Context())

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
