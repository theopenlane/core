package interceptors

import (
	"context"
	"time"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/intercept"
	"github.com/theopenlane/core/internal/ent/generated/jobrunnerregistrationtoken"
)

// InterceptorJobRunnerRegistrationToken is middleware to only list non expired
// tokens
func InterceptorJobRunnerRegistrationToken() ent.Interceptor {
	return intercept.TraverseJobRunnerRegistrationToken(
		func(_ context.Context, q *generated.JobRunnerRegistrationTokenQuery) error {
			q.Where(
				jobrunnerregistrationtoken.Or(
					jobrunnerregistrationtoken.And(
						jobrunnerregistrationtoken.ExpiresAtGT(time.Now()),
						jobrunnerregistrationtoken.JobRunnerIDIsNil(),
					),
					jobrunnerregistrationtoken.JobRunnerIDIsNil(),
				),
			)

			return nil
		},
	)
}
