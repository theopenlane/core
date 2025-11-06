//go:build cli

package register

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
			"registerUser": registerUserHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register register command: %v", err))
	}
}
