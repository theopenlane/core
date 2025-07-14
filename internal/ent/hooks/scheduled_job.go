package hooks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"entgo.io/ent"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/hook"
	"github.com/theopenlane/core/pkg/models"
	"github.com/theopenlane/core/pkg/windmill"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/utils/ulids"
)

// HookScheduledJobCreate verifies a scheduled job has
// a cron and the configuration matches what is expected
// It also validates the download URL and creates a Windmill flow if configured
func HookScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ScheduledJobFunc(func(ctx context.Context,
			mutation *generated.ScheduledJobMutation) (generated.Value, error) {
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

			cron, hasCron := mutation.Cron()

			if err := validateCron(cron, hasCron); err != nil {
				return nil, err
			}

			if mutation.Op() == ent.OpCreate {
				if err := createWindmillFlow(ctx, mutation); err != nil {
					return nil, fmt.Errorf("failed to create windmill flow: %w", err)
				}
			}

			return next.Mutate(ctx, mutation)
		})
	}, ent.OpUpdate|ent.OpUpdateOne|ent.OpCreate)
}

// createWindmillFlow creates a Windmill flow for the scheduled job
func createWindmillFlow(ctx context.Context, mutation *generated.ScheduledJobMutation) error {
	windmillClient := mutation.Client().Windmill
	if windmillClient == nil {
		return nil
	}

	// Get job details
	title, hasTitle := mutation.Title()
	if !hasTitle {
		return errors.New("title is required for windmill flow creation") // nolint:err113
	}

	downloadURL, _ := mutation.DownloadURL()

	if downloadURL == "" {
		return errors.New("download_url is required for windmill flow creation") // nolint:err113
	}

	rawCode, err := downloadRawCode(ctx, downloadURL)
	if err != nil {
		return err
	}

	flowPath := fmt.Sprintf("f/openlane/scheduled_job_%s", generateFlowPath())
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

	resp, err := windmillClient.CreateFlow(ctx, flowReq)
	if err != nil {
		return err
	}

	mutation.SetWindmillPath(resp.Path)
	return nil
}

// downloadRawCode downloads raw code from a URL that will be wrapped into a Windmill flow
func downloadRawCode(ctx context.Context, downloadURL string) (string, error) {
	if downloadURL == "" {
		return "", errors.New("download_url is required for raw code download") // nolint:err113
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download raw code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download raw code, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// generateFlowPath generates a unique random flow path
func generateFlowPath() string {
	return strings.ToLower(ulids.New().String())
}

// HookControlScheduledJobCreate verifies a job that can be attached to a control/subcontrol has
// a cron and the configuration matches what is expected
func HookControlScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ControlScheduledJobFunc(func(ctx context.Context,
			mutation *generated.ControlScheduledJobMutation) (generated.Value, error) {
			cron, hasCron := mutation.Cron()

			if entx.CheckIsSoftDelete(ctx) {
				return next.Mutate(ctx, mutation)
			}

			if err := validateCron(cron, hasCron); err != nil {
				return nil, err
			}

			if mutation.Op() == ent.OpCreate && hasCron && cron != "" {
				if err := createWindmillScheduledJob(ctx, mutation); err != nil {
					return nil, fmt.Errorf("failed to create windmill scheduled job: %w", err)
				}
			}

			return next.Mutate(ctx, mutation)
		})
	}, ent.OpUpdate|ent.OpUpdateOne|ent.OpCreate)
}

func validateCron(cron models.Cron, hasCron bool) error {
	if !hasCron || cron == "" {
		return nil
	}

	return cron.Validate()
}

// createWindmillScheduledJob creates a Windmill scheduled job for the control scheduled job
func createWindmillScheduledJob(ctx context.Context, mutation *generated.ControlScheduledJobMutation) error {
	windmillClient := mutation.Client().Windmill
	if windmillClient == nil {
		return nil
	}

	// Get job details
	jobID, hasJobID := mutation.JobID()
	if !hasJobID {
		return errors.New("job_id is required for windmill scheduled job creation") // nolint:err113
	}

	cron, hasCron := mutation.Cron()
	if !hasCron || cron == "" {
		return errors.New("cron is required for windmill scheduled job creation") // nolint:err113
	}

	entConfig := mutation.Client().EntConfig
	if entConfig == nil {
		return errors.New("ent config is required") // nolint:err113
	}

	scheduledJobPath := fmt.Sprintf("s/openlane/control_scheduled_job_%s", generateFlowPath())

	enabled := true
	scheduleReq := windmill.CreateScheduledJobRequest{
		Path:     scheduledJobPath,
		Schedule: string(cron),
		FlowPath: "", // This would need to be set based on the job configuration
		Enabled:  &enabled,
		Summary:  fmt.Sprintf("Control scheduled job %s", jobID),
	}

	resp, err := windmillClient.CreateScheduledJob(ctx, scheduleReq)
	if err != nil {
		return err
	}

	_ = resp.Path

	return nil
}
