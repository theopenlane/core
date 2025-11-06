//go:build cli

package trustcentercompliance

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
		"openlaneclient.CreateTrustCenterComplianceInput": reflect.TypeOf(openlaneclient.CreateTrustCenterComplianceInput{}),
		"openlaneclient.UpdateTrustCenterComplianceInput": reflect.TypeOf(openlaneclient.UpdateTrustCenterComplianceInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
	}); err != nil {
		panic(fmt.Sprintf("failed to register trustcentercompliance command: %v", err))
	}
}
