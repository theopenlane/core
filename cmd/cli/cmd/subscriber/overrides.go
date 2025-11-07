//go:build cli

package subscribers

import "github.com/theopenlane/core/cmd/cli/internal/speccli"

func init() {
	speccli.RegisterOverride("subscriber", func(spec *speccli.CommandSpec) error {
		// Customize spec-driven behaviour (columns, flags, hooks) here if needed.
		return nil
	})
}
