package hooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"text/template"

	"entgo.io/ent"
	"github.com/google/cel-go/cel"
	"github.com/theopenlane/entx"
	"github.com/theopenlane/iam/auth"

	"github.com/theopenlane/core/internal/ent/entityops"
	"github.com/theopenlane/core/internal/ent/eventqueue"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/standard"
	"github.com/theopenlane/core/internal/ent/generated/task"
	"github.com/theopenlane/core/internal/ent/taskrules"
	"github.com/theopenlane/core/pkg/celx"
	"github.com/theopenlane/core/pkg/gala"
	"github.com/theopenlane/core/pkg/logx"

	"github.com/theopenlane/core/common/enums"
)

const taskRuleSource = "entityops"

// ErrMissingTaskTemplate indicates a rule fired but no taskrules.Template is registered for it
var ErrMissingTaskTemplate = errors.New("entityops: missing task template")

// RegisterGalaTaskRuleListeners registers one gala listener per eligible schema evaluating each schema's rules on mutation and creates suggested
// task records
func RegisterGalaTaskRuleListeners(registry *gala.Registry) ([]gala.ListenerID, error) {
	var ids []gala.ListenerID

	for _, schema := range entityops.TaskRuleEligibleSchemas() {
		listenerIDs, err := gala.RegisterListeners(registry, gala.Definition[eventqueue.MutationGalaPayload]{
			Topic:      eventqueue.MutationTopic(eventqueue.MutationConcernDirect, schema.Name),
			Name:       "taskrules." + schema.Snake,
			Operations: taskRuleOperations(schema),
			Handle: func(ctx gala.HandlerContext, payload eventqueue.MutationGalaPayload) error {
				return handleTaskRuleMutation(ctx, schema, payload)
			},
		})
		if err != nil {
			return nil, err
		}

		ids = append(ids, listenerIDs...)
	}

	return ids, nil
}

// taskRuleOperations returns the mutation operations to subscribe to for schema
func taskRuleOperations(schema *entityops.Schema) []string {
	ops := []string{ent.OpCreate.String()}

	for _, rule := range schema.AllTaskRules() {
		if rule.Rule.Trigger == entx.TaskRuleOnCreateOrUpdate {
			ops = append(ops, ent.OpUpdate.String())
			break
		}
	}

	return ops
}

func handleTaskRuleMutation(ctx gala.HandlerContext, schema *entityops.Schema, payload eventqueue.MutationGalaPayload) error {
	handlerCtx, client, ok := eventqueue.ClientFromHandler(ctx)
	if !ok {
		return nil
	}

	entityID, ok := eventqueue.MutationEntityID(payload, ctx.Envelope.Headers.Properties)
	if !ok || entityID == "" || schema.Load == nil {
		return nil
	}

	systemCtx := taskRuleSystemContext(handlerCtx.Context)

	raw, err := schema.Load(systemCtx, client, entityID)
	if err != nil {
		if generated.IsNotFound(err) {
			return nil
		}

		return err
	}

	entity := map[string]json.RawMessage{}
	if err := json.Unmarshal(raw, &entity); err != nil {
		return err
	}

	placeholders := entityPlaceholders(entityID, entity)

	for _, fieldRule := range schema.AllTaskRules() {
		if !operationAllowed(fieldRule.Rule.Trigger, payload.Operation) {
			continue
		}

		value, ok := ruleValue(fieldRule, entity)
		if !ok {
			continue
		}

		rendered, err := evaluateRule(systemCtx, client, fieldRule.Rule, value, placeholders)
		if err != nil {
			logx.FromContext(systemCtx).Error().Err(err).Str("rule", fieldRule.Rule.RuleID).Msg("entityops: task rule evaluation failed")
			continue
		}

		for _, t := range rendered {
			if err := createSuggestedTask(systemCtx, client, schema, entityID, t); err != nil {
				return err
			}
		}
	}

	return nil
}

// taskRuleSystemContext augments gala's restored caller with the capabilities task-rule
// evaluation needs
// CapOrgSupport is needed to bypass assigner user on task creation
func taskRuleSystemContext(ctx context.Context) context.Context {
	caller, ok := auth.CallerFromContext(ctx)
	if !ok || caller == nil {
		caller = &auth.Caller{}
	}

	return auth.WithCaller(ctx, caller.WithCapabilities(auth.CapInternalOperation|auth.CapOrgSupport))
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

		switch v := decoded.(type) {
		case string:
			out[key] = v
		case float64, bool:
			out[key] = fmt.Sprint(v)
		}
	}

	return out
}

func operationAllowed(trigger entx.TaskRuleTrigger, operation string) bool {
	if trigger == entx.TaskRuleOnCreateOnly {
		return operation == ent.OpCreate.String()
	}

	return operation == ent.OpCreate.String() || operation == ent.OpUpdate.String()
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

	data := make(map[string]any, len(placeholders)+2)
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
	sourceKey := schema.Snake + rendered.Key
	idempotencyKey := fmt.Sprintf("%s:%s:%s%s", taskRuleSource, schema.Snake, entityID, rendered.Key)

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
		if isMissingKeyError(err) {
			return false, nil
		}

		return false, err
	}

	return fire, nil
}

// isMissingKeyError reports whether err is a CEL "no such key" evaluation error
func isMissingKeyError(err error) bool {
	return strings.Contains(err.Error(), "no such key")
}

func evaluateCELList(ctx context.Context, expression string, value any) ([]any, error) {
	evaluator, err := taskRuleEvaluator()
	if err != nil {
		return nil, err
	}

	out, _, err := evaluator.Evaluate(ctx, expression, map[string]any{"value": value})
	if err != nil {
		if isMissingKeyError(err) {
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
		return nil, fmt.Errorf("expression %q did not evaluate to a list", expression)
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
