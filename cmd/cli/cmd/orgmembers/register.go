//go:build cli

package orgmembers

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
		"openlaneclient.CreateOrgMembershipInput": reflect.TypeOf(openlaneclient.CreateOrgMembershipInput{}),
		"openlaneclient.UpdateOrgMembershipInput": reflect.TypeOf(openlaneclient.UpdateOrgMembershipInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks:  map[string]speccli.CreateHookFactory{"orgMembershipCreate": orgMembershipCreateHook},
		UpdateHooks:  map[string]speccli.UpdateHookFactory{"orgMembershipUpdate": orgMembershipUpdateHook},
		DeleteHooks:  map[string]speccli.DeleteHookFactory{"orgMembershipDelete": orgMembershipDeleteHook},
		GetHooks:     map[string]speccli.GetHookFactory{"orgMembershipGet": orgMembershipGetHook},
	}); err != nil {
		panic(fmt.Sprintf("failed to register org members command: %v", err))
	}
}
