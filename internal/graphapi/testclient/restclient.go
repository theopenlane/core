package testclient

import (
	"context"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"

	api "github.com/theopenlane/core/common/openapi"
)

// RestClient is the interface that wraps the openlane API REST client methods
type RestClient interface {
	// Register a new user with the API
	Register(context.Context, *api.RegisterRequest) (*api.RegisterReply, error)
	// Login to the API
	Login(context.Context, *api.LoginRequest) (*api.LoginReply, error)
	// Refresh a user's access token
	Refresh(context.Context, *api.RefreshRequest) (*api.RefreshReply, error)
	// Switch the current organization context
	Switch(context.Context, *api.SwitchOrganizationRequest) (*api.SwitchOrganizationReply, error)
	// VerifyEmail verifies the email address of a user
	VerifyEmail(context.Context, *api.VerifyRequest) (*api.VerifyReply, error)
	// ResendEmail re-sends the verification email to the user
	ResendEmail(context.Context, *api.ResendRequest) (*api.ResendReply, error)
	// ForgotPassword sends a password reset email to the user
	ForgotPassword(context.Context, *api.ForgotPasswordRequest) (*api.ForgotPasswordReply, error)
	// ResetPassword resets the user's password
	ResetPassword(context.Context, *api.ResetPasswordRequest) (*api.ResetPasswordReply, error)
	// AcceptInvite accepts an invite to join an organization
	AcceptInvite(context.Context, *api.InviteRequest) (*api.InviteReply, error)
	// Webfinger retrieves SSO status information via the webfinger endpoint
	Webfinger(context.Context, string) (*api.SSOStatusReply, error)
	// OAuthRegister registers or logs in a user using an OAuth provider
	OAuthRegister(context.Context, *api.OauthTokenRequest) (*api.LoginReply, error)
	// ValidateTOTP validates a user's TOTP or recovery code
	ValidateTOTP(context.Context, *api.TFARequest) (*api.TFAReply, error)
	// AccountAccess checks if a subject has a specific relation to an object
	AccountAccess(context.Context, *api.AccountAccessRequest) (*api.AccountAccessReply, error)
	// AccountRoles lists the relations a subject has in relation to an object
	AccountRoles(context.Context, *api.AccountRolesRequest) (*api.AccountRolesReply, error)
	// AccountRolesOrganization lists roles a user has for an organization
	AccountRolesOrganization(context.Context, *api.AccountRolesOrganizationRequest) (*api.AccountRolesOrganizationReply, error)
	// AccountFeatures lists features a user has for an organization
	AccountFeatures(context.Context, *api.AccountFeaturesRequest) (*api.AccountFeaturesReply, error)
	// RegisterRunner registers a new job runner node with the server
	RegisterRunner(context.Context, *api.JobRunnerRegistrationRequest) (*api.JobRunnerRegistrationReply, error)
}

// NewRestClient creates a new API v1 client that implements the Openlane Client interface
func NewRestClient(config Config, opts ...ClientOption) (_ RestClient, err error) {
	c := &APIv1{
		Config: &config,
	}

	// create the HTTP sling requester if it is not set with the default client
	if c.Requester == nil {
		c.Requester, err = httpsling.New(
			httpsling.Client(
				httpclient.CookieJar(nil), // Use a cookie jar to store session cookies
			),
		)
		if err != nil {
			return nil, err
		}
	}

	// Apply our options
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Ensure the APIv1 implements the OpenlaneClient interface
var _ RestClient = &APIv1{}

// Register a new user with the API
func (s *APIv1) Register(ctx context.Context, in *api.RegisterRequest) (out *api.RegisterReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("register")),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// Login to the API
func (s *APIv1) Login(ctx context.Context, in *api.LoginRequest) (out *api.LoginReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("login")),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newAuthenticationError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// Refresh a user's access token
func (s *APIv1) Refresh(ctx context.Context, in *api.RefreshRequest) (out *api.RefreshReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("refresh")),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newAuthenticationError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// Switch the current organization context
func (s *APIv1) Switch(ctx context.Context, in *api.SwitchOrganizationRequest) (out *api.SwitchOrganizationReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("switch")),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newAuthenticationError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// VerifyEmail verifies the email address of a user
func (s *APIv1) VerifyEmail(ctx context.Context, in *api.VerifyRequest) (out *api.VerifyReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Get(v1Path("verify")),
		httpsling.QueryParam("token", in.Token))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// ResendEmail resends the verification email to the user
func (s *APIv1) ResendEmail(ctx context.Context, in *api.ResendRequest) (out *api.ResendReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("resend")),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// ForgotPassword sends a password reset email to the user
func (s *APIv1) ForgotPassword(ctx context.Context, in *api.ForgotPasswordRequest) (out *api.ForgotPasswordReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("forgot-password")),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// ResetPassword resets the user's password
func (s *APIv1) ResetPassword(ctx context.Context, in *api.ResetPasswordRequest) (out *api.ResetPasswordReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("password-reset")),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// AcceptInvite accepts an invite to join an organization
func (s *APIv1) AcceptInvite(ctx context.Context, in *api.InviteRequest) (out *api.InviteReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Get(v1Path("invite")),
		httpsling.QueryParam("token", in.Token))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// VerifySubscriberEmail verifies the email address of a subscriber
func (s *APIv1) VerifySubscriberEmail(ctx context.Context, in *api.VerifySubscribeRequest) (out *api.VerifySubscribeReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Get(v1Path("subscribe/verify")),
		httpsling.QueryParam("token", in.Token))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// Webfinger retrieves SSO status information via the webfinger endpoint.
func (s *APIv1) Webfinger(ctx context.Context, resource string) (out *api.SSOStatusReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Get("/.well-known/webfinger"),
		httpsling.QueryParam("resource", resource))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// OAuthRegister registers or logs in a user using an OAuth provider.
func (s *APIv1) OAuthRegister(ctx context.Context, in *api.OauthTokenRequest) (out *api.LoginReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post("/oauth/register"),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// ValidateTOTP validates a user's TOTP or recovery code.
func (s *APIv1) ValidateTOTP(ctx context.Context, in *api.TFARequest) (out *api.TFAReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("2fa/validate")),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// AccountAccess checks if a subject has access to an object.
func (s *APIv1) AccountAccess(ctx context.Context, in *api.AccountAccessRequest) (out *api.AccountAccessReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("account/access")),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// AccountRoles lists the relations a subject has in relation to an object.
func (s *APIv1) AccountRoles(ctx context.Context, in *api.AccountRolesRequest) (out *api.AccountRolesReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("account/roles")),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// AccountRolesOrganization lists roles a user has for an organization.
func (s *APIv1) AccountRolesOrganization(ctx context.Context, in *api.AccountRolesOrganizationRequest) (out *api.AccountRolesOrganizationReply, err error) {
	path := v1Path("account/roles/organization")
	if in.ID != "" {
		path += "/" + in.ID
	}

	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Get(path))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// AccountFeatures lists features a user has for an organization.
func (s *APIv1) AccountFeatures(ctx context.Context, in *api.AccountFeaturesRequest) (out *api.AccountFeaturesReply, err error) {
	path := v1Path("account/features")
	if in.ID != "" {
		path += "/" + in.ID
	}

	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Get(path))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Error)
	}

	return out, nil
}

// RegisterRunner registers a new job runner node with the server.
func (s *APIv1) RegisterRunner(ctx context.Context, in *api.JobRunnerRegistrationRequest) (out *api.JobRunnerRegistrationReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("runners")),
		httpsling.Body(in))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return nil, newRequestError(resp.StatusCode, out.Reply.Error)
	}

	return out, nil
}

func v1Path(path string) string {
	return "/v1/" + path
}
