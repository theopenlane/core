package hooks

import (
	"context"
	"fmt"
	"strings"

	"entgo.io/ent"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/jobrunner"
	"github.com/theopenlane/core/internal/ent/generated/jobtemplate"
	"github.com/theopenlane/core/internal/ent/generated/organization"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/privacy/utils"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/windmill"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/utils/ulids"
)

// HookJobTemplate verifies a scheduled job has
// a cron and the configuration matches what is expected
// It also validates the download URL and creates a Windmill flow if configured
func HookJobTemplate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.JobTemplateFunc(func(ctx context.Context,
			mutation *generated.JobTemplateMutation,
		) (generated.Value, error) {
			if entx.CheckIsSoftDeleteType(ctx, mutation.Type()) {
				return next.Mutate(ctx, mutation)
			}

			// validate download URL if provided
			downloadURL, hasDownloadURL := mutation.DownloadURL()
			if hasDownloadURL && downloadURL != "" {
				if _, err := models.ValidateURL(downloadURL); err != nil {
					return nil, fmt.Errorf("invalid download URL: %w", err)
				}
			}

			var hasWindmillUpdate bool

			// check for updates to windmill fields
			platform, _ := mutation.Platform()
			title, _ := mutation.Title()
			description, _ := mutation.Description()

			if downloadURL != "" || platform != "" || title != "" || description != "" {
				hasWindmillUpdate = true
			}

			// capture old values
			var oldWindmillPath, oldDownloadURL string

			if hasWindmillUpdate && mutation.Op().Is(ent.OpUpdateOne) {
				var err error

				oldWindmillPath, err = mutation.OldWindmillPath(ctx)
				if err != nil {
					return nil, err
				}

				oldDownloadURL, err = mutation.OldDownloadURL(ctx)
				if err != nil {
					return nil, err
				}
			}

			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return retVal, err
			}

			switch mutation.Op() {
			case ent.OpCreate:
				flowPath, err := createWindmillFlow(ctx, mutation)
				if err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to create windmill flow")

					return nil, fmt.Errorf("failed to create job template: %w", err)
				}

				// set the flow on the job template
				jobTemplate, ok := retVal.(*generated.JobTemplate)
				if !ok {
					logx.FromContext(ctx).Error().Msg("failed to get job template from return value")

					return retVal, nil
				}

				if err := mutation.Client().JobTemplate.UpdateOneID(jobTemplate.ID).
					SetWindmillPath(flowPath).
					Exec(ctx); err != nil {
					logx.FromContext(ctx).Error().Err(err).Msg("failed to set windmill path on job template")

					return nil, fmt.Errorf("failed to set windmill path on job template: %w", err)
				}
			case ent.OpUpdate, ent.OpUpdateOne:
				if hasWindmillUpdate {
					if err := updateWindmillFlow(ctx, mutation, oldWindmillPath, oldDownloadURL); err != nil {
						logx.FromContext(ctx).Error().Err(err).Msg("failed to update windmill flow")

						return nil, fmt.Errorf("failed to update job template: %w", err)
					}
				}
			}

			return retVal, nil
		})
	}, ent.OpUpdate|ent.OpUpdateOne|ent.OpCreate)
}

// createWindmillFlow creates a Windmill flow for the scheduled job
// returns the path of the flow
func createWindmillFlow(ctx context.Context, mutation *generated.JobTemplateMutation) (string, error) {
	windmillClient := mutation.Client().Windmill
	if windmillClient == nil {
		return "", nil
	}

	title, _ := mutation.Title()
	downloadURL, _ := mutation.DownloadURL()

	rawCode, err := downloadRawCode(ctx, downloadURL)
	if err != nil {
		return "", err
	}

	flowPath, err := generateFlowPath(ctx, mutation)
	if err != nil {
		return "", err
	}

	description, _ := mutation.Description()
	platform, _ := mutation.Platform()

	flowReq := windmill.CreateFlowRequest{
		Path:        flowPath,
		Summary:     title,
		Description: description,
		Value:       []any{rawCode},
		Language:    platform,
	}

	resp, err := windmillClient.CreateFlow(ctx, flowReq)
	if err != nil {
		return "", err
	}

	return resp.Path, nil
}

// updateWindmillFlow updates a Windmill flow for the scheduled job
// this happens after the job template is updated in the database
func updateWindmillFlow(ctx context.Context, mutation *generated.JobTemplateMutation, oldWindmillPath, oldDownloadURL string) error {
	windmillClient := mutation.Client().Windmill
	if windmillClient == nil {
		return nil
	}

	downloadURL, hasDownloadURL := mutation.DownloadURL()

	platform, hasPlatform := mutation.Platform()

	var rawCode string

	if hasDownloadURL && downloadURL != "" && downloadURL != oldDownloadURL {
		var err error

		rawCode, err = downloadRawCode(ctx, downloadURL)
		if err != nil {
			return err
		}
	}

	flowReq := windmill.UpdateFlowRequest{}

	if rawCode != "" {
		flowReq.Value = []any{rawCode}
	}

	if hasPlatform && platform != "" {
		flowReq.Language = platform
	}

	// ensure we have a valid windmill path,
	// this could happen on a bulk update where we couldn't get the path
	// via the OldWindmillPath function
	if oldWindmillPath == "" {
		ids, err := mutation.IDs(ctx)
		if err != nil {
			return err
		}

		if len(ids) == 0 {
			logx.FromContext(ctx).Error().Msg("no ids found, unable to update windmill flow")

			return nil
		}

		jt, err := mutation.Client().JobTemplate.Query().
			Where(jobtemplate.IDIn(ids...)).
			Select(jobtemplate.FieldWindmillPath).
			Only(ctx)
		if err != nil {
			return err
		}

		oldWindmillPath = jt.WindmillPath
	}

	return windmillClient.UpdateFlow(ctx, oldWindmillPath, flowReq)
}

// downloadRawCode downloads raw code from a URL that will be wrapped into a Windmill flow
func downloadRawCode(ctx context.Context, downloadURL string) (string, error) {
	if downloadURL == "" {
		return "", fmt.Errorf("%w: download_url", ErrFieldRequired)
	}

	requestor, err := httpsling.New(
		httpsling.URL(downloadURL),
		httpsling.Get(),
	)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create httpsling requestor")

		return "", ErrInternalServerError
	}

	var rawCode string

	resp, err := requestor.ReceiveWithContext(ctx, &rawCode)
	if err != nil {
		return "", fmt.Errorf("failed to download raw code: %w", err)
	}

	defer resp.Body.Close()

	if !httpsling.IsSuccess(resp) {
		return "", fmt.Errorf("failed to download raw code, status: %d", resp.StatusCode) //nolint:err113
	}

	return rawCode, nil
}

// generateFlowPath generates a unique random flow path based on the ent config
func generateFlowPath(ctx context.Context, mutation utils.GenericMutation) (string, error) {
	entConfig := mutation.Client().EntConfig
	if entConfig == nil {
		logx.FromContext(ctx).Error().Msg("ent config is required, but not set, unable to create scheduled job")

		return "", ErrInternalServerError
	}

	var flowType string
	if _, ok := mutation.(*generated.ScheduledJobMutation); ok {
		flowType = "scheduled_job"
	} else if _, ok := mutation.(*generated.JobTemplateMutation); ok {
		flowType = "job_template"
	} else {
		return "", fmt.Errorf("unknown mutation type: %T", mutation) //nolint:err113
	}

	folderName, err := getCustomerFolderName(ctx, mutation)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to get customer folder name")

		return "", err
	}

	// the `f` here means folder, this is a convention used by Windmill
	// and should most likely always be the case. They also have `u` for users
	// this should create a path like `f/org_name/scheduled_job_12345678901234567890123456789012`
	return fmt.Sprintf("f/%s/%s_%s", folderName, flowType, strings.ToLower(ulids.New().String())), nil
}

func getCustomerFolderName(ctx context.Context, mutation utils.GenericMutation) (string, error) {
	// get organization name
	orgID, err := auth.GetOrganizationIDFromContext(ctx)
	if err != nil {
		return "", err
	}

	// allow to bypass auth checks for org name
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	org, err := mutation.Client().Organization.Query().
		Where(organization.ID(orgID)).
		Select(organization.FieldName).
		Only(allowCtx)
	if err != nil {
		return "", err
	}

	// Windmill requires folder names to be alphanumeric or underscores
	// Replace any invalid characters with underscores
	// see SpecialCharValidator in ent validator to see that the org name can only contain
	// alphanumeric characters, hyphens, underscores, periods, commas, and ampersands
	validName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			return r
		}

		return '_'
	}, org.Name)

	return string(validName), nil
}

// HookScheduledJobCreate verifies a job that can be attached to a control/subcontrol has
// a cron and the configuration matches what is expected
func HookScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ScheduledJobFunc(func(ctx context.Context,
			mutation *generated.ScheduledJobMutation,
		) (generated.Value, error) {
			if entx.CheckIsSoftDeleteType(ctx, mutation.Type()) {
				return next.Mutate(ctx, mutation)
			}

			cron, _ := mutation.Cron()
			jobID, _ := mutation.JobID()

			// if the cron is not set on create, attempt to inherit from the job template
			// let the schema do the validation
			if jobID != "" {
				job, err := mutation.Client().JobTemplate.Query().
					Where(jobtemplate.ID(jobID)).
					Select(jobtemplate.FieldCron).
					Only(ctx)
				if err != nil {
					return nil, err
				}

				if cron == "" && job.Cron != nil {
					cron = *job.Cron
					mutation.SetCron(cron)
				}
			}

			// validate the job runner is in the organization
			jobRunnerID, hasJobRunnerID := mutation.JobRunnerID()
			if hasJobRunnerID && jobRunnerID != "" {
				exists, err := mutation.Client().JobRunner.Query().
					Where(jobrunner.ID(jobRunnerID)).
					Exist(ctx)
				if err != nil {
					return nil, err
				}

				if !exists {
					logx.FromContext(ctx).Debug().Str("job_runner_id", jobRunnerID).Msg("requested job runner not found")

					return nil, &generated.NotFoundError{}
				}
			}

			retVal, err := next.Mutate(ctx, mutation)
			if err != nil {
				return retVal, err
			}

			// TODO: how is the job updated in windmill on update, this hook only handles
			// the creation of the job
			if err := createWindmillScheduledJob(ctx, mutation); err != nil {
				return nil, fmt.Errorf("failed to create scheduled job: %w", err)
			}

			return retVal, nil
		})
	}, ent.OpCreate)
}

// createWindmillScheduledJob creates a scheduled job in Windmill, this should be called after the
// the scheduled job is created in the database, to ensure the fields are set correctly
func createWindmillScheduledJob(ctx context.Context, mutation *generated.ScheduledJobMutation) error {
	windmillClient := mutation.Client().Windmill
	if windmillClient == nil {
		return nil
	}

	jobID, _ := mutation.JobID()
	cron, _ := mutation.Cron()

	scheduledJobPath, err := generateFlowPath(ctx, mutation)
	if err != nil {
		return err
	}

	// get the flow path from the job template
	flowPath, err := mutation.Client().JobTemplate.Query().
		Where(jobtemplate.ID(jobID)).
		Select(jobtemplate.FieldWindmillPath).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to get job template windmill path: %w", err)
	}

	enabled, _ := mutation.Active()

	scheduleReq := windmill.CreateScheduledJobRequest{
		Path:     scheduledJobPath,
		Schedule: string(cron),
		Enabled:  &enabled,
		FlowPath: flowPath.WindmillPath,
		Summary:  fmt.Sprintf("Scheduled Job - %s", jobID),
	}

	_, err = windmillClient.CreateScheduledJob(ctx, scheduleReq)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to create windmill scheduled job")

		return fmt.Errorf("failed to create scheduled job: %w", err)
	}

	return nil
}
