package hooks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"entgo.io/ent"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/httpsling"

	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/internal/ent/generated/jobrunner"
	"github.com/theopenlane/core/internal/ent/generated/jobtemplate"
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
			if entx.CheckIsSoftDelete(ctx) {
				return next.Mutate(ctx, mutation)
			}

			// validate download URL if provided
			downloadURL, hasDownloadURL := mutation.DownloadURL()
			if hasDownloadURL && downloadURL != "" {
				if _, err := models.ValidateURL(downloadURL); err != nil {
					return nil, fmt.Errorf("invalid download URL: %w", err)
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
					log.Error().Err(err).Msg("failed to create windmill flow")

					return nil, fmt.Errorf("failed to create job template: %w", err)
				}

				// set the flow on the job template
				jobTemplate, ok := retVal.(*generated.JobTemplate)
				if !ok {
					log.Error().Msg("failed to get job template from return value")

					return retVal, nil
				}

				if err := mutation.Client().JobTemplate.UpdateOneID(jobTemplate.ID).
					SetWindmillPath(flowPath).
					Exec(ctx); err != nil {
					log.Error().Err(err).Msg("failed to set windmill path on job template")

					return nil, fmt.Errorf("failed to set windmill path on job template: %w", err)
				}
			case ent.OpUpdate, ent.OpUpdateOne:
				if err := updateWindmillFlow(ctx, mutation); err != nil {
					log.Error().Err(err).Msg("failed to update windmill flow")

					return nil, fmt.Errorf("failed to update job template: %w", err)
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

	flowPath, err := generateFlowPath(mutation)
	if err != nil {
		return "", err
	}

	summary := title
	if description, _ := mutation.Description(); description != "" {
		summary = fmt.Sprintf("%s - %s", title, description)
	}

	platform, _ := mutation.Platform()

	flowReq := windmill.CreateFlowRequest{
		Path:     flowPath,
		Summary:  summary,
		Value:    []any{rawCode},
		Language: platform,
	}

	log.Warn().Str("flow_path", flowPath).Str("summary", summary).Msg("creating windmill flow")

	resp, err := windmillClient.CreateFlow(ctx, flowReq)
	if err != nil {
		return "", err
	}

	return resp.Path, nil
}

// updateWindmillFlow updates a Windmill flow for the scheduled job
func updateWindmillFlow(ctx context.Context, mutation *generated.JobTemplateMutation) error {
	windmillClient := mutation.Client().Windmill
	if windmillClient == nil {
		return nil
	}

	oldWindmillPath, err := mutation.OldWindmillPath(ctx)
	if err != nil {
		return err
	}

	oldDownloadURL, err := mutation.OldDownloadURL(ctx)
	if err != nil {
		return err
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

	if hasPlatform {
		flowReq.Language = platform
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
		log.Error().Err(err).Msg("failed to create httpsling requestor")

		return "", ErrInternalServerError
	}

	var rawCode string
	resp, err := requestor.ReceiveWithContext(ctx, &rawCode)
	if err != nil {
		return "", fmt.Errorf("failed to download raw code: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download raw code, status: %d", resp.StatusCode) //nolint:err113
	}

	return rawCode, nil
}

// generateFlowPath generates a unique random flow path based on the ent config
func generateFlowPath(mutation GenericMutation) (string, error) {
	entConfig := mutation.Client().EntConfig
	if entConfig == nil {
		log.Error().Msg("ent config is required, but not set, unable to create scheduled job")

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

	// TODO: should this path be the organization name instead of everything into `openlane` folder
	return fmt.Sprintf("s/%s/%s_%s", entConfig.Windmill.FolderName, flowType, strings.ToLower(ulids.New().String())), nil
}

// HookScheduledJobCreate verifies a job that can be attached to a control/subcontrol has
// a cron and the configuration matches what is expected
func HookScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ScheduledJobFunc(func(ctx context.Context,
			mutation *generated.ScheduledJobMutation,
		) (generated.Value, error) {
			if entx.CheckIsSoftDelete(ctx) {
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
					log.Debug().Str("job_runner_id", jobRunnerID).Msg("requested job runner not found")

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

	scheduledJobPath, err := generateFlowPath(mutation)
	if err != nil {
		return err
	}

	enabled := true
	scheduleReq := windmill.CreateScheduledJobRequest{
		Path:     scheduledJobPath,
		Schedule: string(cron),
		Enabled:  &enabled,
		Summary:  fmt.Sprintf("scheduled job - %s", jobID),
	}

	_, err = windmillClient.CreateScheduledJob(ctx, scheduleReq)

	return err
}
