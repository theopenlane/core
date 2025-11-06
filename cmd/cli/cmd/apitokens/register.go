//go:build cli

package apitokens

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
		"openlaneclient.CreateAPITokenInput": reflect.TypeOf(openlaneclient.CreateAPITokenInput{}),
		"openlaneclient.UpdateAPITokenInput": reflect.TypeOf(openlaneclient.UpdateAPITokenInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
	}); err != nil {
		panic(fmt.Sprintf("failed to register api token command: %v", err))
	}
}
