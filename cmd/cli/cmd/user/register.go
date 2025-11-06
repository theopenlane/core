//go:build cli

package user

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
		"openlaneclient.UpdateUserInput": reflect.TypeOf(openlaneclient.UpdateUserInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		UpdateHooks: map[string]speccli.UpdateHookFactory{
			"userUpdate": updateUserHook,
		},
		GetHooks: map[string]speccli.GetHookFactory{
			"userGet": getUserHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register user command: %v", err))
	}
}
