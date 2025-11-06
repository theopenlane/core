//go:build cli

package dnsverification

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
		"openlaneclient.CreateDNSVerificationInput": reflect.TypeOf(openlaneclient.CreateDNSVerificationInput{}),
		"openlaneclient.UpdateDNSVerificationInput": reflect.TypeOf(openlaneclient.UpdateDNSVerificationInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
	}); err != nil {
		panic(fmt.Sprintf("failed to register dnsverification command: %v", err))
	}
}
