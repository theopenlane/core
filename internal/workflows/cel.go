package workflows

import (
	"github.com/google/cel-go/cel"
)

// NewCelEnv builds the workflow CEL environment using the provided options parameters
func NewCelEnv(opts ...ConfigOpts) (*cel.Env, error) {
	cfg := NewDefaultConfig(opts...)

	return NewCELEnvWithConfig(cfg)
}

// NewCELEnv builds the workflow CEL environment using the provided config
func NewCELEnvWithConfig(cfg *Config) (*cel.Env, error) {
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
		cel.Variable("assignments", cel.DynType),
		cel.Variable("instance", cel.DynType),
		cel.Variable("initiator", cel.StringType),
		// its safe not to check these because NewConfig sets defaults
		cel.ParserRecursionLimit(cfg.CEL.ParserRecursionLimit),
		cel.ParserExpressionSizeLimit(cfg.CEL.ParserExpressionSizeLimit),
		cel.ASTValidators(cel.ValidateComprehensionNestingLimit(cfg.CEL.ComprehensionNestingLimit)),
		cel.CrossTypeNumericComparisons(cfg.CEL.CrossTypeNumericComparisons),
	}

	if cfg.CEL.IdentifierEscapeSyntax {
		envOpts = append(envOpts, cel.EnableIdentifierEscapeSyntax())
	}

	if cfg.CEL.ExtendedValidations {
		envOpts = append(envOpts, cel.ExtendedValidations())
	}

	if cfg.CEL.OptionalTypes {
		envOpts = append(envOpts, cel.OptionalTypes())
	}

	if cfg.CEL.MacroCallTracking {
		envOpts = append(envOpts, cel.EnableMacroCallTracking())
	}

	env, err := cel.NewEnv(envOpts...)
	if err != nil {
		return nil, ErrFailedToBuildCELEnv
	}

	return env, nil
}
