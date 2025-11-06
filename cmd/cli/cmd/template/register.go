//go:build cli

package templates

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
		"openlaneclient.CreateTemplateInput": reflect.TypeOf(openlaneclient.CreateTemplateInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks: map[string]speccli.CreateHookFactory{
			"templateCreate": templateCreateHook,
		},
		GetHooks: map[string]speccli.GetHookFactory{
			"templateGet": templateGetHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register template command: %v", err))
	}
}
