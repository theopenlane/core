package token

import (
	"context"

	"github.com/theopenlane/utils/contextx"
)

type JobRunnerRegistrationToken struct {
	PrivacyToken
	token string
}

func NewJobRunnerRegistrationToken(token string) JobRunnerRegistrationToken {
	return JobRunnerRegistrationToken{
		token: token,
	}
}

func (token *JobRunnerRegistrationToken) GetToken() string {
	return token.token
}

func (t *JobRunnerRegistrationToken) SetToken(token string) {
	t.token = token
}

func NewContextWithJobRunnerRegistrationToken(parent context.Context, token string) context.Context {
	return contextx.With(parent, &JobRunnerRegistrationToken{
		token: token,
	})
}

func JobRunnerRegistrationTokenFromContext(ctx context.Context) *JobRunnerRegistrationToken {
	token, ok := contextx.From[*JobRunnerRegistrationToken](ctx)
	if !ok {
		return nil
	}

	return token
}
