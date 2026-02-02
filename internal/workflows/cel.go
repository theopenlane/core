package workflows

import (
	"github.com/google/cel-go/cel"

	"github.com/theopenlane/core/pkg/celx"
)

// CELExpressionScope determines which variables are available when compiling CEL expressions
type CELExpressionScope int

const (
	// CELScopeBase includes the base workflow variables only
	CELScopeBase CELExpressionScope = iota
	// CELScopeAction includes action-specific variables (assignments, instance, initiator)
	CELScopeAction
)

// NewCELEnv builds the workflow CEL environment using the provided config and scope
func NewCELEnv(cfg *Config, scope CELExpressionScope) (*cel.Env, error) {
	if cfg == nil {
		cfg = NewDefaultConfig()
	}

	envOpts := []cel.EnvOption{
		cel.Variable("object", cel.DynType),
		cel.Variable("user_id", cel.StringType),
		cel.Variable("changed_fields", cel.ListType(cel.StringType)),
		cel.Variable("changed_edges", cel.ListType(cel.StringType)),
		cel.Variable("added_ids", cel.MapType(cel.StringType, cel.ListType(cel.StringType))),
		cel.Variable("removed_ids", cel.MapType(cel.StringType, cel.ListType(cel.StringType))),
		cel.Variable("event_type", cel.StringType),
		cel.Variable("proposed_changes", cel.DynType),
	}

	if scope == CELScopeAction {
		envOpts = append(envOpts,
			cel.Variable("assignments", cel.DynType),
			cel.Variable("instance", cel.DynType),
			cel.Variable("initiator", cel.StringType),
		)
	}

	env, err := celx.NewEnv(envConfigFrom(cfg), envOpts...)
	if err != nil {
		return nil, ErrFailedToBuildCELEnv
	}

	return env, nil
}

// envConfigFrom converts a workflows Config to a celx EnvConfig
func envConfigFrom(cfg *Config) celx.EnvConfig {
	return celx.EnvConfig{
		ParserRecursionLimit:        cfg.CEL.ParserRecursionLimit,
		ParserExpressionSizeLimit:   cfg.CEL.ParserExpressionSizeLimit,
		ComprehensionNestingLimit:   cfg.CEL.ComprehensionNestingLimit,
		ExtendedValidations:         cfg.CEL.ExtendedValidations,
		OptionalTypes:               cfg.CEL.OptionalTypes,
		IdentifierEscapeSyntax:      cfg.CEL.IdentifierEscapeSyntax,
		CrossTypeNumericComparisons: cfg.CEL.CrossTypeNumericComparisons,
		MacroCallTracking:           cfg.CEL.MacroCallTracking,
	}
}
