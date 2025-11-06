//go:build cli

package orgsubscription

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
		GetHooks: map[string]speccli.GetHookFactory{
			"orgSubscriptionGet": orgSubscriptionGetHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register orgsubscription command: %v", err))
	}
}
