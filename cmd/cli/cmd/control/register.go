//go:build cli

package control

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
		"openlaneclient.CreateControlInput": reflect.TypeOf(openlaneclient.CreateControlInput{}),
		"openlaneclient.UpdateControlInput": reflect.TypeOf(openlaneclient.UpdateControlInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks: map[string]speccli.CreateHookFactory{
			"controlCreate": controlCreateHook,
		},
		UpdateHooks: map[string]speccli.UpdateHookFactory{
			"controlUpdate": controlUpdateHook,
		},
		GetHooks: map[string]speccli.GetHookFactory{
			"controlGet": controlGetHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register control command: %v", err))
	}
}
