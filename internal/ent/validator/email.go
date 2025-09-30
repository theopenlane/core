package validator

import (
	"context"
	"errors"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/pkg/metrics"
)

// EmailVerificationConfig is the configuration for email verification
type EmailVerificationConfig struct {
	// Enabled indicates whether email verification is enabled
	Enabled bool `json:"enabled" koanf:"enabled" default:"false" description:"enable email verification"`
	// EnableAutoUpdateDisposable indicates whether to automatically update disposable email addresses
	EnableAutoUpdateDisposable bool `json:"enableAutoUpdateDisposable" koanf:"enableAutoUpdateDisposable" default:"true" description:"automatically update disposable email addresses"`
	// AllowedEmailTypes indicates which types of email addresses are allowed
	AllowedEmailTypes AllowedEmailTypes `json:"allowedEmailTypes" koanf:"allowedEmailTypes"`
}

// AllowedEmailTypes defines the allowed email types for verification
type AllowedEmailTypes struct {
	// Disposable indicates whether disposable email addresses are allowed
	Disposable bool `json:"disposable" koanf:"disposable" default:"false" description:"allow disposable email addresses"`
	// Free indicates whether free email addresses are allowed
	Free bool `json:"free" koanf:"free" default:"false" description:"allow free email addresses"`
	// Role indicates whether role-based email addresses are allowed
	Role bool `json:"role" koanf:"role" default:"false" description:"allow role-based email addresses such as info@, support@"`
}

const (
	// ResultDisposable indicates the email is disposable
	ResultDisposable = "disposable"
	// ResultFree indicates the email is from a free provider
	ResultFree = "free"
	// ResultRole indicates the email is a role-based account, e.g. info@, support@
	ResultRole = "role"
	// ResultSyntax indicates the email has invalid syntax
	ResultSyntax = "syntax"
)

var (
	ErrEmailNotAllowed = errors.New("email address is not allowed, please use your corporate email address")
)

// VerifyEmailAddress verifies the given email address based on the configuration
// and returns whether it is verified, the verification result, and any error encountered
func (c *EmailVerificationConfig) VerifyEmailAddress(ctx context.Context, email string) (bool, *emailverifier.Result, error) {
	verifier := emailverifier.NewVerifier()

	ret, err := verifier.Verify(email)
	if err != nil {
		log.Error().Err(err).Msg("error verifying email address")

		return false, nil, err
	}

	if !ret.Syntax.Valid {
		log.Warn().Str("email", email).Msg("email address failed syntax check")
		metrics.RecordEmailVerification(false, ResultSyntax)

		return false, nil, nil
	}

	if ret.Disposable && !c.AllowedEmailTypes.Disposable {
		log.Warn().Str("email", email).Msg("email address is disposable and disposable emails are not allowed")
		metrics.RecordEmailVerification(false, ResultDisposable)

		return false, nil, nil
	}

	if ret.Free && !c.AllowedEmailTypes.Free {
		log.Warn().Str("email", email).Msg("email address is free and free emails are not allowed")
		metrics.RecordEmailVerification(false, ResultFree)

		return false, nil, nil
	}

	if ret.RoleAccount && !c.AllowedEmailTypes.Role {
		log.Warn().Str("email", email).Msg("email address is role-based and role-based emails are not allowed")
		metrics.RecordEmailVerification(false, ResultRole)

		return false, nil, nil
	}

	return true, ret, nil
}
