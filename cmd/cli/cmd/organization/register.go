//go:build cli

package organization

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
		"openlaneclient.CreateOrganizationInput": reflect.TypeOf(openlaneclient.CreateOrganizationInput{}),
		"openlaneclient.UpdateOrganizationInput": reflect.TypeOf(openlaneclient.UpdateOrganizationInput{}),
		"openlaneclient.OrganizationWhereInput":  reflect.TypeOf(openlaneclient.OrganizationWhereInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks: map[string]speccli.CreateHookFactory{
			"organizationCreate": createOrganizationHook,
		},
		UpdateHooks: map[string]speccli.UpdateHookFactory{
			"organizationUpdate": updateOrganizationHook,
		},
		GetHooks: map[string]speccli.GetHookFactory{
			"organizationGet": getOrganizationHook,
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register organization command: %v", err))
	}
}
