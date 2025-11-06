//go:build cli

package search

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
			"globalSearch": globalSearchHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register search command: %v", err))
	}
}
