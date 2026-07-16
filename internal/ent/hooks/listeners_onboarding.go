package hooks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"github.com/samber/lo"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	onboardingent "github.com/theopenlane/core/internal/ent/generated/onboarding"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/internal/onboarding"
	"github.com/theopenlane/core/pkg/gala"
)

const onboardingTaskSource = "onboarding"

// onboardingTaskKeyReplacer replaces frameworks such as "soc 2" to become "soc-2"
var onboardingTaskKeyReplacer = strings.NewReplacer(" ", "-", "/", "-", "_", "-", ":", "-")

type suggestedTask struct {
	Key         string
	Title       string
	Details     string
	Priority    int
	AvailableAt *models.DateTime
	Metadata    map[string]any
}

// RegisterGalaOnboardingListeners registers onboarding mutation listeners on Gala.
func RegisterGalaOnboardingListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	return gala.RegisterListeners(registry, gala.Definition[eventqueue.MutationGalaPayload]{
		Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, generated.TypeOnboarding),
		Name:       "onboarding.suggested_tasks",
		Operations: []string{ent.OpCreate.String()},
		Handle:     handleOnboardingCreated,
	})
}

func handleOnboardingCreated(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
	ctx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	id, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || id == "" {
		return nil
	}

	allowCtx := privacy.DecisionContext(ctx.Context, privacy.Allow)

	record, err := client.Onboarding.Query().
		Where(onboardingent.IDEQ(id)).
		Only(allowCtx)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	return createOnboardingTasks(ctx.Context, client, record.OrganizationID, lo.Assign(map[string]any{}, record.CompanyDetails, record.UserDetails, record.Compliance))
}

func createOnboardingTasks(ctx context.Context, client *generated.Client, orgID string, answers map[string]any) error {
	if client == nil || orgID == "" {
		return nil
	}

	questionnaire, err := onboarding.Catalog(ctx, client)
	if err != nil {
		return err
	}

	tasks := []suggestedTask{}
	now := time.Now()

	for _, step := range questionnaire.Steps {
		for _, rule := range step.Tasks {
			tasks = append(tasks, generateTaskFromRule(rule, nil, now))
		}

		for _, question := range step.Questions {
			answer, ok := answers[question.Key]
			if !ok || len(question.Tasks) == 0 {
				continue
			}

			for match, rule := range question.Tasks {
				// run the same logic on all members of the tasks
				if match == "eachSelected" {
					values, _ := retrieveValueFromAnswer[[]string](answer)
					for _, value := range values {
						tasks = append(tasks, generateTaskFromRule(rule, onboardingTaskTemplateValues(value, getOptionLabel(question.Options, value)), now))
					}

					continue
				}

				value, _ := retrieveValueFromAnswer[string](answer)
				if value == match {
					tasks = append(tasks, generateTaskFromRule(rule, onboardingTaskTemplateValues(match, getOptionLabel(question.Options, match)), now))
				}
			}
		}
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	for _, t := range tasks {
		idempotencyKey := fmt.Sprintf("%s:%s:%s", onboardingTaskSource, orgID, t.Key)

		exists, err := client.Task.Query().
			Where(
				task.OwnerIDEQ(orgID),
				task.IdempotencyKeyEQ(idempotencyKey),
				task.DeletedAtIsNil(),
			).
			Exist(allowCtx)
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		mutation := client.Task.Create().
			SetOwnerID(orgID).
			SetTitle(t.Title).
			SetDetails(t.Details).
			SetSystemGenerated(true).
			SetIsSuggested(true).
			SetPriority(t.Priority).
			SetSource(onboardingTaskSource).
			SetSourceKey(t.Key).
			SetIdempotencyKey(idempotencyKey)

		if t.AvailableAt != nil {
			mutation.SetAvailableAt(*t.AvailableAt)
		}

		// if len(t.Metadata) > 0 {
		// 	mutation.SetMetadata(t.Metadata)
		// }

		if _, err := mutation.Save(allowCtx); err != nil {
			return err
		}
	}

	return nil
}

func generateTaskFromRule(rule models.TaskRule, values map[string]string, t time.Time) suggestedTask {
	task := suggestedTask{
		Key:      replaceOnboardingTemplateHolder(rule.Key, values),
		Title:    replaceOnboardingTemplateHolder(rule.Title, values),
		Details:  replaceOnboardingTemplateHolder(rule.Details, values),
		Priority: rule.Priority,
		Metadata: rule.Metadata,
	}

	if rule.AvailableAfterDays > 0 {
		availableAt := models.DateTime(t.Add(time.Duration(rule.AvailableAfterDays) * 24 * time.Hour))
		task.AvailableAt = &availableAt
	}

	return task
}

func replaceOnboardingTemplateHolder(value string, replacements map[string]string) string {
	for key, replacement := range replacements {
		value = strings.ReplaceAll(value, "{"+key+"}", replacement)
	}

	return value
}

func onboardingTaskTemplateValues(value, label string) map[string]string {
	if label == "" {
		label = value
	}

	return map[string]string{
		"key":   onboardingTaskKeyReplacer.Replace(strings.ToLower(strings.TrimSpace(value))),
		"value": value,
		"label": label,
	}
}

func retrieveValueFromAnswer[T string | []string](value any) (T, bool) {
	var zero T

	switch any(zero).(type) {
	case string:
		switch v := value.(type) {
		case bool:
			return any(fmt.Sprint(v)).(T), true
		case string:
			return any(v).(T), true
		default:
			return any(fmt.Sprint(v)).(T), true
		}
	case []string:
		switch v := value.(type) {
		case []string:
			return any(v).(T), true
		case []any:
			out := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok && s != "" {
					out = append(out, s)
				}
			}
			return any(out).(T), true
		case string:
			if v != "" {
				return any([]string{v}).(T), true
			}
		}
	}

	return zero, false
}

func getOptionLabel(options []models.QuestionOption, value string) string {
	for _, option := range options {
		if option.Value == value {
			return option.Label
		}
	}

	return ""
}
