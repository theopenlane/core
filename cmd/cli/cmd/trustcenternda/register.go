//go:build cli

package trustcenternda

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
		"openlaneclient.CreateTrustCenterNDAInput": reflect.TypeOf(openlaneclient.CreateTrustCenterNDAInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks:  map[string]speccli.CreateHookFactory{"trustCenterNdaCreate": trustCenterNdaCreateHook},
		UpdateHooks:  map[string]speccli.UpdateHookFactory{"trustCenterNdaUpdate": trustCenterNdaUpdateHook},
	}); err != nil {
		panic(fmt.Sprintf("failed to register trust center NDA command: %v", err))
	}

	cmd := findTrustCenterNdaCommand()
	if cmd == nil {
		panic("failed to locate registered trust center NDA command")
	}

	attachTrustCenterNdaExtras(cmd)
}
