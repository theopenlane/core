package validator

import (
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
	EnableAutoUpdateDisposable bool `json:"enableautoupdatedisposable" koanf:"enableautoupdatedisposable" default:"true" description:"automatically update disposable email addresses"`
	// EnableGravatarCheck indicates whether to check for Gravatar existence
	EnableGravatarCheck bool `json:"enablegravatarcheck" koanf:"enablegravatarcheck" default:"true" description:"check for Gravatar existence"`
	// EnableSMTPCheck indicates whether to check email by smtp
	EnableSMTPCheck bool `json:"enablesmtpcheck" koanf:"enablesmtpcheck" default:"false" description:"check email by smtp"`
	// AllowedEmailTypes defines the allowed email types for verification
	AllowedEmailTypes AllowedEmailTypes `json:"allowedemailtypes" koanf:"allowedemailtypes"`
}

// EmailVerifier is a wrapper around the emailverifier.Verifier with additional configuration
type EmailVerifier struct {
	// Client is the email verifier client
	Client *emailverifier.Verifier
	// AllowedEmailTypes defines the allowed email types for verification
	AllowedEmailTypes AllowedEmailTypes
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
	// ResultUnknown indicates the email verification result is unknown, usually due to an error during verification
	ResultUnknown = "unknown"
	// ResultValid indicates the email is valid
	ResultValid = "valid"
)

var (
	ErrEmailNotAllowed = errors.New("email address is not allowed, please use your corporate email address")
)

func (c *EmailVerificationConfig) NewVerifier() *EmailVerifier {
	if c == nil || !c.Enabled {
		return nil
	}

	v := emailverifier.NewVerifier()

	if c.EnableGravatarCheck {
		v.EnableGravatarCheck()
	}

	if c.EnableSMTPCheck {
		v.EnableSMTPCheck()
	}

	if c.EnableAutoUpdateDisposable {
		v.EnableAutoUpdateDisposable()
	}

	return &EmailVerifier{
		Client:            v,
		AllowedEmailTypes: c.AllowedEmailTypes,
	}
}

// VerifyEmailAddress verifies the given email address based on the configuration
// and returns whether it is verified, the verification result, and any error encountered
func (c *EmailVerifier) VerifyEmailAddress(email string) (bool, *emailverifier.Result, error) {
	if c == nil || c.Client == nil {
		return true, nil, nil
	}

	ret, err := c.Client.Verify(email)
	if err != nil {
		log.Error().Err(err).Msg("error verifying email address")
		metrics.RecordEmailValidation(false, ResultUnknown)

		return false, nil, err
	}

	if !ret.Syntax.Valid {
		log.Warn().Str("email", email).Msg("email address failed syntax check")
		metrics.RecordEmailValidation(false, ResultSyntax)

		return false, nil, nil
	}

	if ret.Disposable && !c.AllowedEmailTypes.Disposable {
		log.Warn().Str("email", email).Msg("email address is disposable and disposable emails are not allowed")
		metrics.RecordEmailValidation(false, ResultDisposable)

		return false, nil, nil
	}

	if ret.Free && !c.AllowedEmailTypes.Free {
		log.Warn().Str("email", email).Msg("email address is free and free emails are not allowed")
		metrics.RecordEmailValidation(false, ResultFree)

		return false, nil, nil
	}

	if ret.RoleAccount && !c.AllowedEmailTypes.Role {
		log.Warn().Str("email", email).Msg("email address is role-based and role-based emails are not allowed")
		metrics.RecordEmailValidation(false, ResultRole)

		return false, nil, nil
	}

	metrics.RecordEmailValidation(true, ResultValid)

	return true, ret, nil
}

// IsFreeDomain checks if the domain of the given email is a free email domain
// If there is no client configured, it returns false
func (c *EmailVerifier) IncludesFreeDomain(domains []string) bool {
	if c == nil || c.Client == nil {
		return false
	}

	for _, domain := range domains {
		if c.Client.IsFreeDomain(domain) {
			return true
		}
	}

	return false
}
