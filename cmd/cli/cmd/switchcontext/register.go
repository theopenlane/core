//go:build cli

package switchcontext

import (
	"embed"
	"fmt"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
)

//go:embed spec.json
var specFS embed.FS

func init() {
	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		Parsers: speccli.DefaultParsers(),
		PrimaryHooks: map[string]speccli.PrimaryHookFactory{
			"switchContext": switchContextHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register switchcontext command: %v", err))
	}
}
