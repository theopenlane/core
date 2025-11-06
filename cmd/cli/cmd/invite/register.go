//go:build cli

package invite

import (
	"embed"
	"fmt"
	"reflect"

	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

//go:embed spec.json
var specFS embed.FS

func init() {
	typeMap := map[string]reflect.Type{
		"openlaneclient.CreateInviteInput": reflect.TypeOf(openlaneclient.CreateInviteInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks:  map[string]speccli.CreateHookFactory{"inviteCreate": inviteCreateHook},
		DeleteHooks:  map[string]speccli.DeleteHookFactory{"inviteDelete": inviteDeleteHook},
		PrimaryHooks: map[string]speccli.PrimaryHookFactory{"inviteAccept": inviteAcceptHook},
	}); err != nil {
		panic(fmt.Sprintf("failed to register invite command: %v", err))
	}

	cmd := findInviteCommand()
	if cmd == nil {
		panic("failed to locate registered invite command")
	}

	attachInviteExtras(cmd)
}
