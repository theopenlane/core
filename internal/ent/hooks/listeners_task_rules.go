package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"sync"
	"text/template"

	"github.com/google/cel-go/cel"
	"github.com/samber/lo"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/notification"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/ent/generated/task"
	"github.com/theopenlane/core/v2/internal/ent/taskrules"
	"github.com/theopenlane/core/v2/pkg/celx"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/jsonx"
	"github.com/theopenlane/logx"

	"github.com/theopenlane/core/common/enums"
)

// init registers the task rule listeners so gala setup picks them up automatically
func init() { registerListeners(TaskRuleListeners) }

// TaskRuleListeners evaluates schema task rules on mutation and creates suggested tasks
func TaskRuleListeners() []gala.Registration {
	return lo.Map(entityops.TaskRuleEligibleSchemas(), func(schema *entityops.Schema, _ int) gala.Registration {
		listener := entityops.MutationListener{
			Schema:     schema,
			Operations: taskRuleOperations(schema),
			Caller: func(restored *auth.Caller, _ entityops.MutationPayload) *auth.Caller {
				return restored.WithCapabilities(auth.CapInternalOperation | auth.CapOrgSupport)
			},
			Handle: handleTaskRuleMutation,
		}

		if schema.Name == generated.TypeOrganization {
			listener.Match = []entityops.FieldMatch{{Field: organization.FieldPersonalOrg, In: []string{"true"}, Negate: true}}
		}

		return listener
	})
}

// taskRuleOperations returns the mutation operations to subscribe to for schema
func taskRuleOperations(schema *entityops.Schema) []string {
	ops := []string{entityops.OpCreate}

	for _, rule := range schema.AllTaskRules() {
		if rule.Rule.Trigger == entx.TaskRuleOnCreateOrUpdate {
			ops = append(ops, entityops.OpUpdate, entityops.OpUpdateOne)
			break
		}
	}

	return ops
}

// handleTaskRuleMutation evaluates the mutated schema's task rules against the loaded entity
// and persists any suggested tasks the rules render
func handleTaskRuleMutation(inv entityops.Invocation, payload entityops.MutationPayload) error {
	raw, err := inv.Schema.Load(inv.Context, inv.Client, inv.EntityID)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	entity, err := jsonx.Decode[map[string]json.RawMessage](raw)
	if err != nil {
		return err
	}

	placeholders := entityPlaceholders(inv.EntityID, entity)

	var tasksCreated int

	for _, fieldRule := range inv.Schema.AllTaskRules() {
		if !operationAllowed(fieldRule.Rule.Trigger, payload.Operation) {
			continue
		}

		value, ok := ruleValue(fieldRule, entity)
		if !ok {
			continue
		}

		rendered, err := evaluateRule(inv.Context, inv.Client, fieldRule.Rule, value, placeholders)
		if err != nil {
			logx.FromContext(inv.Context).Error().Err(err).Str("rule", fieldRule.Rule.RuleID).Msg("entityops: task rule evaluation failed")

			continue
		}

		for _, t := range rendered {
			if err := createSuggestedTask(inv.Context, inv.Client, inv.Schema, inv.EntityID, t); err != nil {
				return err
			}

			tasksCreated++
		}
	}

	// add notification when organization create tasks are created, this allows
	// the frontend to wait for this even to load the dashboard for the user
	if inv.Schema.Name == generated.TypeOrganization && payload.Operation == entityops.OpCreate && tasksCreated > 0 {
		if err := notifyOrganizationReady(inv.Context, inv.Client, inv.EntityID); err != nil {
			return err
		}
	}

	return nil
}

// notifyOrganizationReady emits NotificationTopicOrganizationReady once an orgs
// suggested tasks have all been created
func notifyOrganizationReady(ctx context.Context, client *generated.Client, organizationID string) error {
	const organizationReadyObjectType = "organization.ready"

	// gala delivers at-least-once, so skip if this org already has the notification
	// if another organization event fails, they all get retriggered and we don't want multiple
	// notifications
	exists, err := client.Notification.Query().
		Where(
			notification.OwnerIDEQ(organizationID),
			notification.ObjectTypeEQ(organizationReadyObjectType),
		).
		Exist(ctx)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	_, err = client.Notification.Create().
		SetOwnerID(organizationID).
		SetNotificationType(enums.NotificationTypeOrganization).
		SetObjectType(organizationReadyObjectType).
		SetTitle("Organization ready").
		SetBody("Your organization is ready. Review your recommended next steps to get started.").
		SetTopic(enums.NotificationTopicOrganizationReady).
		Save(ctx)

	return err
}

// ruleValue resolves what "value" binds to in a rule's CEL expression
func ruleValue(fieldRule entityops.FieldTaskRule, entity map[string]json.RawMessage) (any, bool) {
	if fieldRule.Field == "" {
		whole := make(map[string]any, len(entity))

		for key, raw := range entity {
			var decoded any
			if json.Unmarshal(raw, &decoded) == nil {
				whole[key] = decoded
			}
		}

		return whole, true
	}

	raw, ok := entity[fieldRule.Field]
	if !ok {
		return nil, true
	}

	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, false
	}

	return decoded, true
}

// entityPlaceholders builds the {fieldname} substitutions available to every rule on this
// entity
func entityPlaceholders(entityID string, entity map[string]json.RawMessage) map[string]string {
	out := map[string]string{"id": entityID}

	for key, raw := range entity {
		var decoded any
		if err := json.Unmarshal(raw, &decoded); err != nil {
			continue
		}

		mergeScalarField(out, key, decoded)
	}

	return out
}

// mergeScalarField sets dst[key] to value's string form when value is a string, float64, or
// bool, and leaves dst untouched for any other type (objects, arrays, null)
func mergeScalarField(dst map[string]string, key string, value any) {
	switch v := value.(type) {
	case string:
		dst[key] = v
	case float64, bool:
		dst[key] = fmt.Sprint(v)
	}
}

func operationAllowed(trigger entx.TaskRuleTrigger, operation string) bool {
	if trigger == entx.TaskRuleOnCreateOnly {
		return operation == entityops.OpCreate
	}

	return lo.Contains([]string{entityops.OpCreate, entityops.OpUpdate, entityops.OpUpdateOne}, operation)
}

// renderedTask is one fully-resolved suggested task, ready to persist
type renderedTask struct {
	Key          string
	Title        string
	Details      string
	Priority     int
	Source       string
	TaskKindName string
	Metadata     map[string]any
}

// evaluateRule evaluates one rule against value: for EachElement rules it expands into one
// renderedTask per list element, resolving {label} via any registered resolver; for Expression
// rules it fires at most one renderedTask when the condition is true. placeholders carries the
// firing entity's own fields (e.g. {id}, {body}), available regardless of which case fires
func evaluateRule(ctx context.Context, client *generated.Client, rule entityops.TaskRuleDescriptor, value any, placeholders map[string]string) ([]renderedTask, error) {
	tmpl, ok := taskrules.Lookup(rule.RuleID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMissingTaskTemplate, rule.RuleID)
	}

	if rule.EachElement != "" {
		elements, err := evaluateCELList(ctx, rule.EachElement, value)
		if err != nil {
			return nil, err
		}

		if m, ok := value.(map[string]any); ok {
			merged := make(map[string]string, len(placeholders)+len(m))
			maps.Copy(merged, placeholders)

			for k, v := range m {
				mergeScalarField(merged, k, v)
			}

			placeholders = merged
		}

		out := make([]renderedTask, 0, len(elements))

		for _, element := range elements {
			elementValue := fmt.Sprint(element)
			label := resolveLabel(ctx, client, rule.RuleID, elementValue)

			rendered, err := renderTask(tmpl, rule.RuleID, elementValue, label, placeholders)
			if err != nil {
				return nil, err
			}

			out = append(out, rendered)
		}

		return out, nil
	}

	fire, err := evaluateCELBool(ctx, rule.Expression, value)
	if err != nil {
		return nil, err
	}

	if !fire {
		return nil, nil
	}

	rendered, err := renderTask(tmpl, rule.RuleID, "", "", placeholders)
	if err != nil {
		return nil, err
	}

	return []renderedTask{rendered}, nil
}

// renderTask executes the template's title, details, and any string metadata values as Go
// templates, with "value"/"label" (EachElement expansion; empty for plain Expression rules) and
// every entity placeholder (e.g. "id", "body") bound as template data -- e.g.
// {{if eq .value "soc2"}}...{{else}}...{{end}} or {{.label}}
func renderTask(tmpl taskrules.Template, ruleID, value, label string, placeholders map[string]string) (renderedTask, error) {
	key := "-" + ruleID
	if value != "" {
		key += "-" + slugifyTaskKey(value)
	}

	data := make(map[string]any, len(placeholders)+2) //nolint:mnd
	for name, v := range placeholders {
		data[name] = v
	}

	data["value"] = value
	data["label"] = label

	title, err := renderTemplateString("title", tmpl.Title, data)
	if err != nil {
		return renderedTask{}, err
	}

	details, err := renderTemplateString("details", tmpl.Details, data)
	if err != nil {
		return renderedTask{}, err
	}

	metadata, err := renderMetadata(tmpl.Metadata, data)
	if err != nil {
		return renderedTask{}, err
	}

	source := string(tmpl.Source)
	if source == "" {
		source = string(taskrules.SourceRecommendations)
	}

	return renderedTask{
		Key:          key,
		Title:        title,
		Details:      details,
		Priority:     tmpl.Priority,
		Source:       source,
		TaskKindName: tmpl.TaskKindName,
		Metadata:     metadata,
	}, nil
}

// renderTemplateString executes text as a Go template against data
func renderTemplateString(name, text string, data map[string]any) (string, error) {
	if text == "" {
		return "", nil
	}

	tmpl, err := template.New(name).Option("missingkey=zero").Parse(text)
	if err != nil {
		return "", fmt.Errorf("parse %s template: %w", name, err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute %s template: %w", name, err)
	}

	return strings.ReplaceAll(buf.String(), "<no value>", ""), nil
}

// renderMetadata applies replacer to every string value in metadata, leaving other value types untouched
func renderMetadata(metadata map[string]any, data map[string]any) (map[string]any, error) {
	if len(metadata) == 0 {
		return metadata, nil
	}

	out := make(map[string]any, len(metadata))

	for key, value := range metadata {
		s, ok := value.(string)
		if !ok {
			out[key] = value
			continue
		}

		rendered, err := renderTemplateString("metadata."+key, s, data)
		if err != nil {
			return nil, err
		}

		out[key] = rendered
	}

	return out, nil
}

var taskKeySlugReplacer = strings.NewReplacer(" ", "-", "/", "-", ":", "-")

func slugifyTaskKey(value string) string {
	return taskKeySlugReplacer.Replace(strings.ToLower(strings.TrimSpace(value)))
}

func createSuggestedTask(ctx context.Context, client *generated.Client, schema *entityops.Schema, entityID string, rendered renderedTask) error {

	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil || caller.OrganizationID == "" {
		return generated.ErrPermissionDenied
	}

	sourceKey := schema.Snake + rendered.Key
	idempotencyKey := fmt.Sprintf("%s:%s:%s%s", "entityops", schema.Snake, entityID, rendered.Key)

	exists, err := client.Task.Query().
		Where(
			task.IdempotencyKeyEQ(idempotencyKey),
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
		SetOwnerID(caller.OrganizationID).
		SetTitle(rendered.Title).
		SetDetails(rendered.Details).
		SetSystemGenerated(true).
		SetIsSuggested(true).
		SetPriority(rendered.Priority).
		SetSource(rendered.Source).
		SetSourceKey(sourceKey).
		SetIdempotencyKey(idempotencyKey)

	if rendered.TaskKindName != "" {
		mutation.SetTaskKindName(rendered.TaskKindName)
	}

	if len(rendered.Metadata) > 0 {
		mutation.SetMetadata(rendered.Metadata)
	}

	_, err = mutation.Save(ctx)

	return err
}

var (
	taskRuleCELOnce      sync.Once
	taskRuleCELEvaluator *celx.Evaluator
	taskRuleCELErr       error
)

// taskRuleEvaluator lazily builds the shared CEL evaluator for task rules: a single "value"
// variable bound to whatever field or whole-entity data a rule is evaluated against
func taskRuleEvaluator() (*celx.Evaluator, error) {
	taskRuleCELOnce.Do(func() {
		env, err := celx.NewEnv(celx.StrictEnvConfig(), cel.Variable("value", cel.DynType))
		if err != nil {
			taskRuleCELErr = err
			return
		}

		taskRuleCELEvaluator = celx.NewEvaluator(env, celx.FastEvalConfig())
	})

	return taskRuleCELEvaluator, taskRuleCELErr
}

func evaluateCELBool(ctx context.Context, expression string, value any) (bool, error) {
	if expression == "" {
		return true, nil
	}

	evaluator, err := taskRuleEvaluator()
	if err != nil {
		return false, err
	}

	fire, err := evaluator.EvaluateBool(ctx, expression, map[string]any{"value": value})
	if err != nil {
		if celx.IsMissingKey(err) {
			return false, nil
		}

		return false, err
	}

	return fire, nil
}

func evaluateCELList(ctx context.Context, expression string, value any) ([]any, error) {
	evaluator, err := taskRuleEvaluator()
	if err != nil {
		return nil, err
	}

	out, _, err := evaluator.Evaluate(ctx, expression, map[string]any{"value": value})
	if err != nil {
		if celx.IsMissingKey(err) {
			return nil, nil
		}

		return nil, err
	}

	decoded, err := celx.ToJSON(out)
	if err != nil {
		return nil, err
	}

	list, ok := decoded.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrExpressionNotList, expression)
	}

	return list, nil
}

// TaskLabelResolver resolves a human-readable label for one EachElement value filling the {label} placeholder in a task template
type TaskLabelResolver func(ctx context.Context, client *generated.Client, value string) string

var taskLabelResolvers = map[string]TaskLabelResolver{
	taskrules.RuleFramework: resolveFrameworkLabel,
}

// resolveLabel looks up value's label via any resolver registered for ruleID, falling back to
// value itself when none is registered or the resolver comes up empty
func resolveLabel(ctx context.Context, client *generated.Client, ruleID, value string) string {
	resolver, ok := taskLabelResolvers[ruleID]
	if !ok {
		return value
	}

	if label := resolver(ctx, client, value); label != "" {
		return label
	}

	return value
}

// resolveFrameworkLabel resolves a framework code (the value submitted for the "frameworks"
// onboarding question, see internal/onboarding/catalog.go's getFrameworkOptions) to its display name
func resolveFrameworkLabel(ctx context.Context, client *generated.Client, value string) string {
	if client == nil {
		return value
	}

	std, err := client.Standard.Query().
		Where(
			standard.FrameworkEQ(value),
			standard.StatusEQ(enums.StandardActive),
		).
		First(ctx)
	if err != nil {
		return value
	}

	if std.ShortName != "" {
		return std.ShortName
	}

	return std.Name
}
