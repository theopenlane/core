//go:build cli

package jobresult

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
		"openlaneclient.CreateJobResultInput": reflect.TypeOf(openlaneclient.CreateJobResultInput{}),
		"openlaneclient.UpdateJobResultInput": reflect.TypeOf(openlaneclient.UpdateJobResultInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
	}); err != nil {
		panic(fmt.Sprintf("failed to register jobresult command: %v", err))
	}
}
