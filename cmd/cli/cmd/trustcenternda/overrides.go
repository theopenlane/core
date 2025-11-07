//go:build cli

package trustcenternda

import "github.com/theopenlane/core/cmd/cli/internal/speccli"

func init() {
	speccli.RegisterOverride("trust-center-nda", func(spec *speccli.CommandSpec) error {
		// Customize spec-driven behaviour (columns, defaults, etc.) here if needed.
		return nil
	})
}
