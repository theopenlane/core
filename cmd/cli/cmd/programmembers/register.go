//go:build cli

package programmembers

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
		"openlaneclient.CreateProgramMembershipInput": reflect.TypeOf(openlaneclient.CreateProgramMembershipInput{}),
		"openlaneclient.UpdateProgramMembershipInput": reflect.TypeOf(openlaneclient.UpdateProgramMembershipInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks:  map[string]speccli.CreateHookFactory{"programMembershipCreate": programMembershipCreateHook},
		UpdateHooks:  map[string]speccli.UpdateHookFactory{"programMembershipUpdate": programMembershipUpdateHook},
		DeleteHooks:  map[string]speccli.DeleteHookFactory{"programMembershipDelete": programMembershipDeleteHook},
		GetHooks:     map[string]speccli.GetHookFactory{"programMembershipGet": programMembershipGetHook},
	}); err != nil {
		panic(fmt.Sprintf("failed to register program members command: %v", err))
	}
}
