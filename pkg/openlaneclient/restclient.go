package openlaneclient

import (
	"context"

	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/httpsling/httpclient"

	"github.com/theopenlane/core/pkg/models"
)

// OpenlaneRestClient is the interface that wraps the openlane API REST client methods
type OpenlaneRestClient interface {
	// Register a new user with the API
	Register(context.Context, *models.RegisterRequest) (*models.RegisterReply, error)
	// Login to the API
	Login(context.Context, *models.LoginRequest) (*models.LoginReply, error)
	// Refresh a user's access token
	Refresh(context.Context, *models.RefreshRequest) (*models.RefreshReply, error)
	// Switch the current organization context
	Switch(context.Context, *models.SwitchOrganizationRequest) (*models.SwitchOrganizationReply, error)
	// VerifyEmail verifies the email address of a user
	VerifyEmail(context.Context, *models.VerifyRequest) (*models.VerifyReply, error)
	// ResendEmail re-sends the verification email to the user
	ResendEmail(context.Context, *models.ResendRequest) (*models.ResendReply, error)
	// ForgotPassword sends a password reset email to the user
	ForgotPassword(context.Context, *models.ForgotPasswordRequest) (*models.ForgotPasswordReply, error)
	// ResetPassword resets the user's password
	ResetPassword(context.Context, *models.ResetPasswordRequest) (*models.ResetPasswordReply, error)
	// AcceptInvite accepts an invite to join an organization
	AcceptInvite(context.Context, *models.InviteRequest) (*models.InviteReply, error)
}

// New creates a new API v1 client that implements the Openlane Client interface
func NewRestClient(config Config, opts ...ClientOption) (_ OpenlaneRestClient, err error) {
	c := &APIv1{
		Config: config,
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
var _ OpenlaneRestClient = &APIv1{}

// Register a new user with the API
func (s *APIv1) Register(ctx context.Context, in *models.RegisterRequest) (out *models.RegisterReply, err error) {
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
func (s *APIv1) Login(ctx context.Context, in *models.LoginRequest) (out *models.LoginReply, err error) {
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
func (s *APIv1) Refresh(ctx context.Context, in *models.RefreshRequest) (out *models.RefreshReply, err error) {
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
func (s *APIv1) Switch(ctx context.Context, in *models.SwitchOrganizationRequest) (out *models.SwitchOrganizationReply, err error) {
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
func (s *APIv1) VerifyEmail(ctx context.Context, in *models.VerifyRequest) (out *models.VerifyReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("verify")),
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

// ResendEmail resends the verification email to the user
func (s *APIv1) ResendEmail(ctx context.Context, in *models.ResendRequest) (out *models.ResendReply, err error) {
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
func (s *APIv1) ForgotPassword(ctx context.Context, in *models.ForgotPasswordRequest) (out *models.ForgotPasswordReply, err error) {
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
func (s *APIv1) ResetPassword(ctx context.Context, in *models.ResetPasswordRequest) (out *models.ResetPasswordReply, err error) {
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
func (s *APIv1) AcceptInvite(ctx context.Context, in *models.InviteRequest) (out *models.InviteReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("invite")),
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

// VerifySubscriberEmail verifies the email address of a subscriber
func (s *APIv1) VerifySubscriberEmail(ctx context.Context, in *models.VerifySubscribeRequest) (out *models.VerifySubscribeReply, err error) {
	resp, err := s.Requester.ReceiveWithContext(ctx, &out,
		httpsling.Post(v1Path("subscribe/verify")),
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

func v1Path(path string) string {
	return "/v1/" + path
}
