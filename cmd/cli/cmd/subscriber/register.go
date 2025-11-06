//go:build cli

package subscribers

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
		CreateHooks: map[string]speccli.CreateHookFactory{
			"subscriberCreate": subscriberCreateHook,
		},
		UpdateHooks: map[string]speccli.UpdateHookFactory{
			"subscriberUpdate": subscriberUpdateHook,
		},
		DeleteHooks: map[string]speccli.DeleteHookFactory{
			"subscriberDelete": subscriberDeleteHook,
		},
		GetHooks: map[string]speccli.GetHookFactory{
			"subscriberGet": subscriberGetHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register subscriber command: %v", err))
	}
}
