package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// maxTemplateRenderDepth limits recursion depth for template rendering to prevent stack overflow
const maxTemplateRenderDepth = 64

// ErrTemplateRenderDepthExceeded is returned when template rendering exceeds the maximum recursion depth
var ErrTemplateRenderDepthExceeded = errors.New("template render depth exceeded")

// renderTemplateValue renders template expressions within a structured value
func renderTemplateValue(ctx context.Context, evaluator *CELEvaluator, input any, vars map[string]any) (any, error) {
	return renderTemplateValueWithDepth(ctx, evaluator, input, vars, 0)
}

// renderTemplateValueWithDepth renders template expressions with depth tracking to prevent stack overflow
func renderTemplateValueWithDepth(ctx context.Context, evaluator *CELEvaluator, input any, vars map[string]any, depth int) (any, error) {
	if depth > maxTemplateRenderDepth {
		return nil, ErrTemplateRenderDepthExceeded
	}

	switch typed := input.(type) {
	case string:
		return renderTemplateString(ctx, evaluator, typed, vars)
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			rendered, err := renderTemplateValueWithDepth(ctx, evaluator, value, vars, depth+1)
			if err != nil {
				return nil, err
			}
			out[key] = rendered
		}
		return out, nil
	case []any:
		out := make([]any, len(typed))
		for i, value := range typed {
			rendered, err := renderTemplateValueWithDepth(ctx, evaluator, value, vars, depth+1)
			if err != nil {
				return nil, err
			}
			out[i] = rendered
		}
		return out, nil
	default:
		return input, nil
	}
}

// renderTemplateText renders a template string and coerces it to text
func renderTemplateText(ctx context.Context, evaluator *CELEvaluator, input string, vars map[string]any) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", nil
	}

	rendered, err := renderTemplateString(ctx, evaluator, input, vars)
	if err != nil {
		return "", err
	}

	return formatTemplateValue(rendered), nil
}

// renderTemplateString renders a template string returning the evaluated value
func renderTemplateString(ctx context.Context, evaluator *CELEvaluator, input string, vars map[string]any) (any, error) {
	if input == "" || !strings.Contains(input, "{{") {
		return input, nil
	}

	if expr, ok := singleTemplateExpression(input); ok {
		return renderTemplateExpression(ctx, evaluator, normalizeTemplateExpr(expr), vars)
	}

	var out strings.Builder
	remaining := input
	for {
		start := strings.Index(remaining, "{{")
		if start < 0 {
			out.WriteString(remaining)
			break
		}

		out.WriteString(remaining[:start])
		tail := remaining[start+2:]
		before, after, ok := strings.Cut(tail, "}}")
		if !ok {
			out.WriteString(remaining[start:])
			break
		}

		expr := strings.TrimSpace(before)
		if expr != "" {
			rendered, err := renderTemplateExpression(ctx, evaluator, normalizeTemplateExpr(expr), vars)
			if err != nil {
				return "", err
			}
			out.WriteString(formatTemplateValue(rendered))
		}

		remaining = after
	}

	return out.String(), nil
}

// singleTemplateExpression extracts a lone template expression from a string
func singleTemplateExpression(input string) (string, bool) {
	trimmed := strings.TrimSpace(input)
	if !strings.HasPrefix(trimmed, "{{") || !strings.HasSuffix(trimmed, "}}") {
		return "", false
	}
	if strings.Count(trimmed, "{{") != 1 || strings.Count(trimmed, "}}") != 1 {
		return "", false
	}
	expr := strings.TrimSpace(trimmed[2 : len(trimmed)-2])
	if expr == "" {
		return "", false
	}
	return expr, true
}

// renderTemplateExpression evaluates a CEL expression with provided vars
func renderTemplateExpression(ctx context.Context, evaluator *CELEvaluator, expr string, vars map[string]any) (any, error) {
	if evaluator == nil {
		return nil, ErrExecutorNotAvailable
	}

	return evaluator.EvaluateValue(ctx, expr, vars)
}

// normalizeTemplateExpr normalizes a template expression for evaluation
func normalizeTemplateExpr(expr string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return expr
	}
	if isTemplateIdentifier(expr) {
		if _, ok := templateRootVars[expr]; !ok {
			return "data." + expr
		}
	}
	return expr
}

// isTemplateIdentifier reports whether an expression is a bare identifier
func isTemplateIdentifier(expr string) bool {
	for i, r := range expr {
		if i == 0 {
			if !isIdentifierStart(r) {
				return false
			}
			continue
		}
		if !isIdentifierChar(r) {
			return false
		}
	}
	return expr != ""
}

// isIdentifierStart reports whether r is a valid start character for an identifier
func isIdentifierStart(r rune) bool {
	return r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

// isIdentifierChar reports whether r is a valid identifier character
func isIdentifierChar(r rune) bool {
	return r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}

// templateRootVars enumerates top-level variables available in templates
var templateRootVars = map[string]struct{}{
	"object":           {},
	"user_id":          {},
	"changed_fields":   {},
	"changed_edges":    {},
	"added_ids":        {},
	"removed_ids":      {},
	"event_type":       {},
	"proposed_changes": {},
	"assignments":      {},
	"instance":         {},
	"initiator":        {},
	"instance_id":      {},
	"definition_id":    {},
	"object_id":        {},
	"object_type":      {},
	"action_key":       {},
	"data":             {},
	"true":             {},
	"false":            {},
	"null":             {},
}

// formatTemplateValue formats an evaluated value into a string
func formatTemplateValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case fmt.Stringer:
		return typed.String()
	case []byte:
		return string(typed)
	}

	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}

	return string(encoded)
}
