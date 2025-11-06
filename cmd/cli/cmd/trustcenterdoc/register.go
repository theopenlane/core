//go:build cli

package trustcenterdoc

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
		"openlaneclient.CreateTrustCenterDocInput": reflect.TypeOf(openlaneclient.CreateTrustCenterDocInput{}),
		"openlaneclient.UpdateTrustCenterDocInput": reflect.TypeOf(openlaneclient.UpdateTrustCenterDocInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
	}); err != nil {
		panic(fmt.Sprintf("failed to register trustcenterdoc command: %v", err))
	}
}
