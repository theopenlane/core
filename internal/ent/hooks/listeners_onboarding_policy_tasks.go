package hooks

import (
	"context"
	"embed"
	"fmt"
	"strings"
	"sync"

	"entgo.io/ent"
	"gopkg.in/yaml.v3"

	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/onboarding"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/internal/workflows"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"
)

//go:embed definitions/onboarding_tasks.yaml
var tasksFS embed.FS

const tasksDefinitionPath = "definitions/onboarding_tasks.yaml"

var (
	automatedTaskDefinitionsOnce sync.Once
	automatedTaskDefinitions     []automatedTaskDefinition
	automatedTaskDefinitionsErr  error
)

type automatedTaskDefinition struct {
	Key            string                 `json:"key" yaml:"key"`
	IfKeyValue     automatedTaskCondition `json:"if_key_value" yaml:"if_key_value"`
	Title          string                 `json:"title" yaml:"title"`
	Details        []string               `json:"details" yaml:"details"`
	URLs           []string               `json:"urls,omitempty" yaml:"urls"`
	IdempotencyKey string                 `json:"idempotency_key" yaml:"idempotency_key"`
}

type automatedTaskCondition struct {
	Key   string `json:"key" yaml:"key"`
	Value bool   `json:"value" yaml:"value"`
}

// RegisterGalaOnboardingTaskListeners registers onboarding task creation listeners.
func RegisterGalaOnboardingTaskListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry,
		gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeOnboarding),
			Name:       "onboarding.tasks",
			Operations: []string{ent.OpCreate.String()},
			Handle:     handleOnboardingTaskGala,
		},
	)
}

func handleOnboardingTaskGala(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	onboardingID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || onboardingID == "" {
		return nil
	}

	allowCtx := workflows.AllowContext(ctx.Context)

	record, err := client.Onboarding.Query().
		Where(onboarding.ID(onboardingID)).
		Only(allowCtx)
	if generated.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return err
	}

	if strings.TrimSpace(record.OrganizationID) == "" {
		return nil
	}

	definitions, err := loadTaskDefinitions()
	if err != nil {
		return err
	}

	for _, definition := range definitions {
		if !shouldProcessDefinition(definition, record.Compliance) {
			continue
		}

		if err := createTask(allowCtx, client, record.OrganizationID, definition); err != nil {
			logx.FromContext(ctx.Context).Error().
				Err(err).
				Str("onboarding_id", onboardingID).
				Str("organization_id", record.OrganizationID).
				Str("task_definition", definition.Key).
				Msg("could not create onboarding task")

			return err
		}
	}

	return nil
}

func loadTaskDefinitions() ([]automatedTaskDefinition, error) {
	automatedTaskDefinitionsOnce.Do(func() {
		data, err := tasksFS.ReadFile(tasksDefinitionPath)
		if err != nil {
			automatedTaskDefinitionsErr = err
			return
		}

		if err := yaml.Unmarshal(data, &automatedTaskDefinitions); err != nil {
			automatedTaskDefinitionsErr = err
		}
	})

	return automatedTaskDefinitions, automatedTaskDefinitionsErr
}

func shouldProcessDefinition(definition automatedTaskDefinition, data map[string]any) bool {
	key := strings.TrimSpace(definition.IfKeyValue.Key)
	if key == "" {
		return false
	}

	value, ok := data[key]
	if !ok {
		return false == definition.IfKeyValue.Value
	}

	v, ok := value.(bool)
	return ok && v == definition.IfKeyValue.Value
}

func createTask(ctx context.Context, client *generated.Client, orgID string, definition automatedTaskDefinition) error {
	key := strings.ReplaceAll(definition.IdempotencyKey, "{{organization_id}}", orgID)
	if strings.TrimSpace(key) == "" {
		return fmt.Errorf("onboarding task %q missing idempotency key", definition.Key)
	}

	exists, err := client.Task.Query().
		Where(
			task.OwnerIDEQ(orgID),
			task.IdempotencyKeyEQ(key),
			task.DeletedAtIsNil(),
		).
		Exist(ctx)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	mutation := client.Task.Create().
		SetOwnerID(orgID).
		SetTitle(definition.Title).
		SetDetails(formatAsMarkdownList(definition.Details)).
		SetSystemGenerated(true).
		SetIdempotencyKey(key).
		SetTags([]string{"onboarding", "policies"})

	if len(definition.URLs) > 0 {
		mutation.SetExternalReferenceURL(definition.URLs)
	}

	_, err = mutation.Save(ctx)
	return err
}

// formatAsMarkdownList takes an array and formats it as markdown so the ui can render
func formatAsMarkdownList(details []string) string {
	if len(details) == 0 {
		return ""
	}

	lines := make([]string, 0, len(details))
	for _, detail := range details {
		detail = strings.TrimSpace(detail)
		if detail == "" {
			continue
		}

		lines = append(lines, "- "+detail)
	}

	return strings.Join(lines, "\n")
}
