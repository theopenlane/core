package celx

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
)

const (
	// EntityVarTarget is the CEL variable bound to the candidate entity JSON in entity expressions
	EntityVarTarget = "target"
	// EntityVarSource is the CEL variable bound to the source entity JSON in entity expressions
	EntityVarSource = "source"
	// EntityVarNow is the CEL variable bound to the logical evaluation timestamp
	EntityVarNow = "now"
)

// NativeEntityEvaluator evaluates boolean CEL expressions against typed entity values, binding the
// candidate entity to "target" and an optional source entity to "source" as native CEL struct types.
// Fields are accessed by their json tag (snake_case), so expressions like
// "target.identity_holder_id == source.id" resolve against the bound projection types directly,
// without decoding to a map[string]any. It satisfies the entityops ExpressionEvaluator and
// SourceAwareExpressionEvaluator method sets, so it is a drop-in for SelectTargets evaluation.
type NativeEntityEvaluator struct {
	// targetType is the concrete struct type bound to the "target" variable
	targetType reflect.Type
	// sourceType is the concrete struct type bound to the "source" variable, nil when unused
	sourceType reflect.Type
	// eval is the underlying compiled-and-cached CEL evaluator
	eval *Evaluator
}

// NewNativeEntityEvaluator builds a typed entity evaluator whose "target" (and optional "source")
// variables are the native CEL types of targetType and sourceType. Pass a nil sourceType when the
// evaluator only needs the candidate entity exposed as "target"
func NewNativeEntityEvaluator(envCfg EnvConfig, evalCfg EvalConfig, targetType reflect.Type, sourceType reflect.Type) (*NativeEntityEvaluator, error) {
	targetType = derefType(targetType)

	nativeArgs := []any{targetType, ext.ParseStructTag("json")}

	vars := []cel.EnvOption{
		cel.Variable(EntityVarTarget, cel.ObjectType(objectTypeName(targetType))),
		cel.Variable(EntityVarNow, cel.TimestampType),
	}

	if sourceType != nil {
		sourceType = derefType(sourceType)

		if sourceType != targetType {
			nativeArgs = append(nativeArgs, sourceType)
		}

		vars = append(vars, cel.Variable(EntityVarSource, cel.ObjectType(objectTypeName(sourceType))))
	}

	env, err := NewEnv(envCfg, append(vars, ext.NativeTypes(nativeArgs...))...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCompileFailed, err)
	}

	return &NativeEntityEvaluator{
		targetType: targetType,
		sourceType: sourceType,
		eval:       NewEvaluator(env, evalCfg),
	}, nil
}

// EvaluateBool evaluates the expression against the candidate entity JSON, exposing it as "target"
func (n *NativeEntityEvaluator) EvaluateBool(ctx context.Context, expression string, data json.RawMessage) (bool, error) {
	target, err := decodeNative(data, n.targetType)
	if err != nil {
		return false, err
	}

	return n.eval.EvaluateBool(ctx, expression, map[string]any{EntityVarTarget: target})
}

// EvaluateBoolWithSource evaluates the expression against the target entity JSON with the source
// entity exposed as "source", so selectors like target.identity_holder_id == source.id work
func (n *NativeEntityEvaluator) EvaluateBoolWithSource(ctx context.Context, expression string, targetData, sourceData json.RawMessage) (bool, error) {
	target, err := decodeNative(targetData, n.targetType)
	if err != nil {
		return false, err
	}

	source, err := decodeNative(sourceData, n.sourceType)
	if err != nil {
		return false, err
	}

	return n.eval.EvaluateBool(ctx, expression, map[string]any{EntityVarTarget: target, EntityVarSource: source})
}

// decodeNative unmarshals entity JSON into a pointer to the supplied struct type for native binding
func decodeNative(data json.RawMessage, t reflect.Type) (any, error) {
	value := reflect.New(t).Interface()
	if err := json.Unmarshal(data, value); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEntityDataInvalid, err)
	}

	return value, nil
}

// derefType resolves a pointer type to its element type
func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t
}

// objectTypeName returns the package-qualified CEL type name cel-go's native provider assigns to t
func objectTypeName(t reflect.Type) string {
	return derefType(t).String()
}
