package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/theopenlane/core/common/enums"
	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/pkg/logx"
	echo "github.com/theopenlane/echox"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/rout"
)

// domainScanRateLimitWindow bounds how often a single organization may trigger a domain scan
const domainScanRateLimitWindow = time.Hour

// domainScanRateLimitKeyPrefix namespaces the per-organization cooldown key in redis
const domainScanRateLimitKeyPrefix = "domainscan:ratelimit:"

// domainScanPerformedBy is recorded on Scan records created by this endpoint
const domainScanPerformedBy = "openlane_domain_scan"

// DomainScanHandler queues a domain scan by creating one pending Scan record and returning
// immediately with its ID. Requests are limited to one per hour, with the exception of system admins
func (h *Handler) DomainScanHandler(ctx echo.Context) error {
	req, err := BindAndValidate[models.DomainScanRequest](ctx)
	if err != nil {
		return h.InvalidInput(ctx, err)
	}

	reqCtx := ctx.Request().Context()

	caller, ok := auth.CallerFromContext(reqCtx)
	if !ok || caller == nil {
		return h.Unauthorized(ctx, ErrAuthenticationRequired)
	}

	if err := rule.RequirePaymentMethod()(reqCtx, nil); err != nil && !errors.Is(err, privacy.Skip) {
		return h.Forbidden(ctx, err)
	}

	if !caller.Has(auth.CapSystemAdmin) {
		allowed, err := h.allowDomainScan(reqCtx, caller.OrganizationID)
		if err != nil {
			logx.FromContext(reqCtx).Error().Err(err).Msg("domain scan: failed checking rate limit")

			return h.InternalServerError(ctx, ErrProcessingRequest)
		}

		if !allowed {
			return h.TooManyRequests(ctx, ErrDomainScanRateLimited)
		}
	}

	create := h.DBClient.Scan.Create().
		SetTarget(req.Domain).
		SetScanType(enums.ScanTypeDomain).
		SetStatus(enums.ScanStatusPending).
		SetPerformedBy(domainScanPerformedBy).
		SetMetadata(map[string]any{"forceRefresh": req.ForceRefresh})

	// only set the owner id if its not a system admin, otherwise set system owned
	if !caller.Has(auth.CapSystemAdmin) {
		create.SetOwnerID(caller.OrganizationID)
	}

	scan, err := create.Save(reqCtx)
	if err != nil {
		logx.FromContext(reqCtx).Error().Err(err).Str("domain", req.Domain).Msg("domain scan: failed creating scan record")

		return h.InternalServerError(ctx, ErrProcessingRequest)
	}

	return h.Success(ctx, &models.DomainScanReply{
		Reply:   rout.Reply{Success: true},
		Message: "domain scan queued",
		ScanID:  scan.ID,
	})
}

// allowDomainScan enforces a per-organization cooldown via a redis, returning false when
// the organization has already triggered a scan within domainScanRateLimitWindow.
// When no redis client is configured the limit is not enforced
func (h *Handler) allowDomainScan(ctx context.Context, orgID string) (bool, error) {
	if h.RedisClient == nil {
		return true, nil
	}

	ok, err := h.RedisClient.SetNX(ctx, domainScanRateLimitKeyPrefix+orgID, time.Now().Unix(), domainScanRateLimitWindow).Result()
	if err != nil {
		return false, err
	}

	return ok, nil
}
