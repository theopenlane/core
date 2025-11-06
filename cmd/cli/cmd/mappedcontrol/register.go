//go:build cli

package mappedcontrol

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
		"openlaneclient.CreateMappedControlInput": reflect.TypeOf(openlaneclient.CreateMappedControlInput{}),
		"openlaneclient.UpdateMappedControlInput": reflect.TypeOf(openlaneclient.UpdateMappedControlInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks:  map[string]speccli.CreateHookFactory{"mappedControlCreate": mappedControlCreateHook},
		UpdateHooks:  map[string]speccli.UpdateHookFactory{"mappedControlUpdate": mappedControlUpdateHook},
		DeleteHooks:  map[string]speccli.DeleteHookFactory{"mappedControlDelete": mappedControlDeleteHook},
		GetHooks:     map[string]speccli.GetHookFactory{"mappedControlGet": mappedControlGetHook},
	}); err != nil {
		panic(fmt.Sprintf("failed to register mapped control command: %v", err))
	}
}
