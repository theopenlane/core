//go:build cli

package subcontrol

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
		"openlaneclient.CreateSubcontrolInput": reflect.TypeOf(openlaneclient.CreateSubcontrolInput{}),
		"openlaneclient.UpdateSubcontrolInput": reflect.TypeOf(openlaneclient.UpdateSubcontrolInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks: map[string]speccli.CreateHookFactory{
			"subcontrolCreate": subcontrolCreateHook,
		},
		UpdateHooks: map[string]speccli.UpdateHookFactory{
			"subcontrolUpdate": subcontrolUpdateHook,
		},
		DeleteHooks: map[string]speccli.DeleteHookFactory{
			"subcontrolDelete": subcontrolDeleteHook,
		},
		GetHooks: map[string]speccli.GetHookFactory{
			"subcontrolGet": subcontrolGetHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register subcontrol command: %v", err))
	}
}
