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

// defaultTrustCenterDomain is the fallback domain for trust center URLs when no custom domain is configured
var defaultTrustCenterDomain string

// SetDefaultTrustCenterDomain sets the fallback domain for trust center URLs
func SetDefaultTrustCenterDomain(domain string) {
	defaultTrustCenterDomain = domain
}

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

// trustCenterResolveResult captures the resolved URL and org name from trust center lookup
type trustCenterResolveResult struct {
	URL     string
	OrgName string
}

// resolveTrustCenterAnonURL loads a trust center and generates an anonymous access token URL
func resolveTrustCenterAnonURL(ctx context.Context, req types.OperationRequest, requestID, trustCenterID, email string, buildURL func(*generated.TrustCenter, string) url.URL) (trustCenterResolveResult, error) {
	tc, err := req.DB.TrustCenter.Query().
		Where(trustcenter.IDEQ(trustCenterID)).
		WithCustomDomain().
		WithSetting().
		Only(ctx)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("trust_center_id", trustCenterID).Msg("failed loading trust center for email")
		return trustCenterResolveResult{}, fmt.Errorf("%w: %w", ErrSendFailed, err)
	}

	var orgName string
	if tc.Edges.Setting != nil {
		orgName = tc.Edges.Setting.CompanyName
	}

	duration := req.DB.TokenManager.Config().TrustCenterNDARequestAccessDuration
	baseURL := buildURL(tc, defaultTrustCenterDomain)

	result, err := urlx.GenerateAnonTokenURL(ctx, req.DB.TokenManager, req.DB.Shortlinks, baseURL, urlx.AnonTokenRequest{
		Prefix:    authmanager.AnonTrustCenterJWTPrefix,
		SubjectID: requestID,
		OrgID:     tc.OwnerID,
		Email:     email,
		Duration:  duration,
		ExtraClaims: func(c *tokens.Claims) {
			c.TrustCenterID = trustCenterID
		},
	})
	if err != nil {
		return trustCenterResolveResult{}, fmt.Errorf("%w: %w", ErrSendFailed, err)
	}

	return trustCenterResolveResult{URL: result.URL, OrgName: orgName}, nil
}

// resolveTrustCenterNDARequestFields populates NDAURL and OrgName on the input when empty
func resolveTrustCenterNDARequestFields(ctx context.Context, req types.OperationRequest, input *TrustCenterNDARequestEmail) error {
	if input.NDAURL != "" || input.RequestID == "" || input.TrustCenterID == "" {
		return nil
	}

	result, err := resolveTrustCenterAnonURL(ctx, req, input.RequestID, input.TrustCenterID, input.Email, trustCenterNDAURL)
	if err != nil {
		return err
	}

	input.NDAURL = result.URL
	if input.OrgName == "" {
		input.OrgName = result.OrgName
	}

	return nil
}

// resolveTrustCenterAuthFields populates AuthURL and OrgName on the input when empty
func resolveTrustCenterAuthFields(ctx context.Context, req types.OperationRequest, input *TrustCenterAuthEmail) error {
	if input.AuthURL != "" || input.RequestID == "" || input.TrustCenterID == "" {
		return nil
	}

	result, err := resolveTrustCenterAnonURL(ctx, req, input.RequestID, input.TrustCenterID, input.Email, trustCenterBaseURL)
	if err != nil {
		return err
	}

	input.AuthURL = result.URL
	if input.OrgName == "" {
		input.OrgName = result.OrgName
	}

	return nil
}
