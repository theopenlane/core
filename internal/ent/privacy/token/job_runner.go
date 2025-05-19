package token

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

// JobRunnerRegistrationToken implements the PrivacyToken interface
type JobRunnerRegistrationToken struct {
	PrivacyToken
	token string
}

// NewJobRunnerRegistrationToken creates a new PrivacyToken of type JobRunnerRegistrationToken
func NewJobRunnerRegistrationToken(token string) JobRunnerRegistrationToken {
	return JobRunnerRegistrationToken{
		token: token,
	}
}

// GetToken from the registration token
func (token *JobRunnerRegistrationToken) GetToken() string {
	return token.token
}

// SetToken sets the token
func (token *JobRunnerRegistrationToken) SetToken(s string) {
	token.token = s
}

// NewContextWithJobRunnerRegistrationToken returns a new context with the job runner registration token
func NewContextWithJobRunnerRegistrationToken(parent context.Context, token string) context.Context {
	return contextx.With(parent, &JobRunnerRegistrationToken{
		token: token,
	})
}

// JobRunnerRegistrationTokenFromContext returns the registration token if available from the context
func JobRunnerRegistrationTokenFromContext(ctx context.Context) *JobRunnerRegistrationToken {
	token, ok := contextx.From[*JobRunnerRegistrationToken](ctx)
	if !ok {
		return nil
	}

	return token
}
