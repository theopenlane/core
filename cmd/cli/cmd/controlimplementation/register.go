//go:build cli

package controlimplementation

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
		"openlaneclient.CreateControlImplementationInput": reflect.TypeOf(openlaneclient.CreateControlImplementationInput{}),
		"openlaneclient.UpdateControlImplementationInput": reflect.TypeOf(openlaneclient.UpdateControlImplementationInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
	}); err != nil {
		panic(fmt.Sprintf("failed to register controlimplementation command: %v", err))
	}
}
