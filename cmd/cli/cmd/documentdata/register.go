//go:build cli

package documentdata

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
		"openlaneclient.CreateDocumentDataInput": reflect.TypeOf(openlaneclient.CreateDocumentDataInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks:  map[string]speccli.CreateHookFactory{"documentDataCreate": documentDataCreateHook},
	}); err != nil {
		panic(fmt.Sprintf("failed to register document data command: %v", err))
	}
}
