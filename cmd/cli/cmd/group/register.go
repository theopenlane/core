//go:build cli

package group

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
		"openlaneclient.CreateGroupInput": reflect.TypeOf(openlaneclient.CreateGroupInput{}),
		"openlaneclient.UpdateGroupInput": reflect.TypeOf(openlaneclient.UpdateGroupInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks:  map[string]speccli.CreateHookFactory{"groupCreate": groupCreateHook},
	}); err != nil {
		panic(fmt.Sprintf("failed to register group command: %v", err))
	}
}
