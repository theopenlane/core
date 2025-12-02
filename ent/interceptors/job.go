package interceptors

import (
	"context"
	"time"

	"entgo.io/ent"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/intercept"
	"github.com/theopenlane/ent/generated/jobrunner"
	"github.com/theopenlane/ent/generated/jobrunnerregistrationtoken"
	"github.com/theopenlane/ent/privacy/rule"
)

// InterceptorJobRunnerRegistrationToken is middleware to only list non expired
// tokens
func InterceptorJobRunnerRegistrationToken() ent.Interceptor {
	return intercept.TraverseJobRunnerRegistrationToken(
		func(_ context.Context, q *generated.JobRunnerRegistrationTokenQuery) error {
			q.Where(
				jobrunnerregistrationtoken.And(
					jobrunnerregistrationtoken.ExpiresAtGT(time.Now()),
					jobrunnerregistrationtoken.JobRunnerIDIsNil(),
				),
			)

			return nil
		},
	)
}

// InterceptorJobRunnerFilterSystemOwned makes sure to always filter out
// system owned runners from responses except the request is from an admin
func InterceptorJobRunnerFilterSystemOwned() ent.Interceptor {
	return intercept.TraverseJobRunner(
		func(ctx context.Context, q *generated.JobRunnerQuery) error {
			isAdmin, err := rule.CheckIsSystemAdminWithContext(ctx)
			if err != nil {
				return err
			}

			if isAdmin {
				return nil
			}

			q.Where(jobrunner.OwnerIDNotNil())

			return nil
		},
	)
}
