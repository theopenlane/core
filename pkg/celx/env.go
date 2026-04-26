package celx

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
)

// NewEnv builds a CEL environment from the provided config and variable declarations
func NewEnv(cfg EnvConfig, vars ...cel.EnvOption) (*cel.Env, error) {
	envOpts := make([]cel.EnvOption, 0, len(vars)+8) //nolint:mnd
	envOpts = append(envOpts, vars...)

	envOpts = append(envOpts,
		cel.ParserRecursionLimit(cfg.ParserRecursionLimit),
		cel.ParserExpressionSizeLimit(cfg.ParserExpressionSizeLimit),
		cel.CrossTypeNumericComparisons(cfg.CrossTypeNumericComparisons),
		// add extensions for parsing
		ext.Strings(),
		cel.StdLib(),
	)

	if cfg.ComprehensionNestingLimit > 0 {
		envOpts = append(envOpts, cel.ASTValidators(cel.ValidateComprehensionNestingLimit(cfg.ComprehensionNestingLimit)))
	}

	if cfg.IdentifierEscapeSyntax {
		envOpts = append(envOpts, cel.EnableIdentifierEscapeSyntax())
	}

	if cfg.ExtendedValidations {
		envOpts = append(envOpts, cel.ExtendedValidations())
	}

	if cfg.OptionalTypes {
		envOpts = append(envOpts, cel.OptionalTypes())
	}

	if cfg.MacroCallTracking {
		envOpts = append(envOpts, cel.EnableMacroCallTracking())
	}

	return cel.NewEnv(envOpts...)
}
