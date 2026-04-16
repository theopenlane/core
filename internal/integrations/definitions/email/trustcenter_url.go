package email

import (
	"context"
	"fmt"
	"net/url"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/trustcenter"
	"github.com/theopenlane/core/internal/httpserve/authmanager"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/domain"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/urlx"
)

// trustCenterBaseURL builds the base URL for a trust center using its custom
// domain when available, falling back to the configured default domain
func trustCenterBaseURL(tc *generated.TrustCenter, defaultDomain string) url.URL {
	u := url.URL{Scheme: "https"}

	if tc.Edges.CustomDomain != nil {
		host := tc.Edges.CustomDomain.CnameRecord
		if normalized, err := domain.NormalizeHostname(host); err == nil {
			host = normalized
		}

		u.Host = host

		return u
	}

	host := defaultDomain
	if normalized, err := domain.NormalizeHostname(host); err == nil {
		host = normalized
	}

	u.Host = host

	return u
}

// trustCenterNDAURL builds the NDA signing URL for a trust center
func trustCenterNDAURL(tc *generated.TrustCenter, defaultDomain string) url.URL {
	u := trustCenterBaseURL(tc, defaultDomain)
	u.Path = "/" + tc.Slug + "/access/sign-nda"

	return u
}

// resolveTrustCenterNDARequestFields loads the trust center from the database and
// populates NDAURL and OrgName on the input when they are empty
func resolveTrustCenterNDARequestFields(ctx context.Context, req types.OperationRequest, client *EmailClient, input *TrustCenterNDARequestEmail) error {
	if input.NDAURL != "" {
		return nil
	}

	if input.RequestID == "" || input.TrustCenterID == "" {
		return nil
	}

	tc, err := req.DB.TrustCenter.Query().
		Where(trustcenter.IDEQ(input.TrustCenterID)).
		WithCustomDomain().
		WithSetting().
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", input.TrustCenterID).Msg("failed loading trust center for NDA email")
		return fmt.Errorf("%w: %w", ErrSendFailed, err)
	}

	if input.OrgName == "" && tc.Edges.Setting != nil {
		input.OrgName = tc.Edges.Setting.CompanyName
	}

	duration := req.DB.TokenManager.Config().TrustCenterNDARequestAccessDuration

	baseURL := trustCenterNDAURL(tc, client.Config.TrustCenterDomain)

	result, err := urlx.GenerateAnonTokenURL(ctx, req.DB.TokenManager, req.DB.Shortlinks, baseURL, urlx.AnonTokenRequest{
		Prefix:    authmanager.AnonTrustCenterJWTPrefix,
		SubjectID: input.RequestID,
		OrgID:     tc.OwnerID,
		Email:     input.Email,
		Duration:  duration,
		ExtraClaims: func(c *tokens.Claims) {
			c.TrustCenterID = input.TrustCenterID
		},
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSendFailed, err)
	}

	input.NDAURL = result.URL

	return nil
}

// resolveTrustCenterAuthFields loads the trust center from the database and
// populates AuthURL and OrgName on the input when they are empty
func resolveTrustCenterAuthFields(ctx context.Context, req types.OperationRequest, client *EmailClient, input *TrustCenterAuthEmail) error {
	if input.AuthURL != "" {
		return nil
	}

	if input.RequestID == "" || input.TrustCenterID == "" {
		return nil
	}

	tc, err := req.DB.TrustCenter.Query().
		Where(trustcenter.IDEQ(input.TrustCenterID)).
		WithCustomDomain().
		WithSetting().
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", input.TrustCenterID).Msg("failed loading trust center for auth email")
		return fmt.Errorf("%w: %w", ErrSendFailed, err)
	}

	if input.OrgName == "" && tc.Edges.Setting != nil {
		input.OrgName = tc.Edges.Setting.CompanyName
	}

	duration := req.DB.TokenManager.Config().TrustCenterNDARequestAccessDuration

	baseURL := trustCenterBaseURL(tc, client.Config.TrustCenterDomain)

	result, err := urlx.GenerateAnonTokenURL(ctx, req.DB.TokenManager, req.DB.Shortlinks, baseURL, urlx.AnonTokenRequest{
		Prefix:    authmanager.AnonTrustCenterJWTPrefix,
		SubjectID: input.RequestID,
		OrgID:     tc.OwnerID,
		Email:     input.Email,
		Duration:  duration,
		ExtraClaims: func(c *tokens.Claims) {
			c.TrustCenterID = input.TrustCenterID
		},
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSendFailed, err)
	}

	input.AuthURL = result.URL

	return nil
}
