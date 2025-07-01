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
	"gopkg.in/yaml.v3"
)

// HookScheduledJobCreate verifies a scheduled job has
// a cadence set or a cron and the configuration matches what is expected
// It also validates the download URL and creates a Windmill flow if configured
func HookScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ScheduledJobFunc(func(ctx context.Context,
			mutation *generated.ScheduledJobMutation) (generated.Value, error) {
			if entx.CheckIsSoftDelete(ctx) {
				return next.Mutate(ctx, mutation)
			}

			// Validate download URL if provided
			downloadURL, hasDownloadURL := mutation.DownloadURL()
			if hasDownloadURL && downloadURL != "" {
				if _, err := models.ValidateURL(downloadURL); err != nil {
					return nil, fmt.Errorf("invalid download URL: %w", err)
				}
			}

			cadence, hasCadence := mutation.Cadence()
			cron, hasCron := mutation.Cron()

			if err := validateCadenceOrCron(&cadence, hasCadence, cron, hasCron); err != nil {
				return nil, err
			}

			// Create Windmill flow if client is available and we're creating a new job
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

	flowValue, err := getFlowValue(ctx, downloadURL)
	if err != nil {
		return err
	}

	flowPath := fmt.Sprintf("f/openlane/scheduled_job_%s", generateFlowPath(title))
	summary := title
	if description, _ := mutation.Description(); description != "" {
		summary = fmt.Sprintf("%s - %s", title, description)
	}

	flowReq := windmill.CreateFlowRequest{
		Path:    flowPath,
		Summary: summary,
		Value:   flowValue,
	}

	resp, err := windmillClient.CreateFlow(ctx, flowReq)
	if err != nil {
		return err
	}

	mutation.SetWindmillPath(resp.Path)
	return nil
}

// getFlowValue returns the appropriate flow value based on downloadURL
func getFlowValue(ctx context.Context, downloadURL string) ([]any, error) {
	if downloadURL == "" {
		return nil, errors.New("download_url is required for windmill flow creation") // nolint:err113
	}

	return downloadFlowDefinition(ctx, downloadURL)
}

// downloadFlowDefinition downloads and parses a Windmill flow definition from a URL
func downloadFlowDefinition(ctx context.Context, downloadURL string) ([]any, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download flow definition: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download flow definition, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var flowDef map[string]any
	if err := yaml.Unmarshal(body, &flowDef); err != nil {
		return nil, fmt.Errorf("failed to parse flow definition YAML: %w", err)
	}

	if value, ok := flowDef["value"]; ok {
		if modules, ok := value.([]any); ok {
			return modules, nil
		}
		if valueMap, ok := value.(map[string]any); ok {
			if modules, ok := valueMap["modules"].([]any); ok {
				return modules, nil
			}
		}
	}

	if modules, ok := flowDef["modules"].([]any); ok {
		return modules, nil
	}

	return nil, fmt.Errorf("invalid flow definition format: missing modules/value array")
}

// generateFlowPath generates a unique random flow path
func generateFlowPath(title string) string {
	return strings.ToLower(ulids.New().String())
}

// HookControlScheduledJobCreate verifies a job that can be attached to a control/subcontrol has
// a cadence set or a cron and the configuration matches what is expected
func HookControlScheduledJobCreate() ent.Hook {
	return hook.On(func(next ent.Mutator) ent.Mutator {
		return hook.ControlScheduledJobFunc(func(ctx context.Context,
			mutation *generated.ControlScheduledJobMutation) (generated.Value, error) {
			cadence, hasCadence := mutation.Cadence()
			cron, hasCron := mutation.Cron()

			if entx.CheckIsSoftDelete(ctx) {
				return next.Mutate(ctx, mutation)
			}

			if err := validateCadenceOrCron(&cadence, hasCadence, cron, hasCron); err != nil {
				return nil, err
			}

			return next.Mutate(ctx, mutation)
		})
	}, ent.OpUpdate|ent.OpUpdateOne|ent.OpCreate)
}

func validateCadenceOrCron(cadence *models.JobCadence, hasCadence bool, cron models.Cron, hasCron bool) error {
	if !hasCadence && (!hasCron || cron == "") {
		return nil
	}

	if hasCadence && hasCron {
		return ErrEitherCadenceOrCron
	}

	if hasCadence {
		if err := cadence.Validate(); err != nil {
			return fmt.Errorf("cadence: %w", err) // nolint:err113
		}
	}

	if hasCron {
		return cron.Validate()
	}

	return nil
}
