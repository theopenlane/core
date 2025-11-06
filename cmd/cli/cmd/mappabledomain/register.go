//go:build cli

package mappabledomain

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
		"openlaneclient.CreateMappableDomainInput": reflect.TypeOf(openlaneclient.CreateMappableDomainInput{}),
		"openlaneclient.UpdateMappableDomainInput": reflect.TypeOf(openlaneclient.UpdateMappableDomainInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
	}); err != nil {
		panic(fmt.Sprintf("failed to register mappabledomain command: %v", err))
	}
}
