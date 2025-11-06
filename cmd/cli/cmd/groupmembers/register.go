//go:build cli

package groupmembers

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
		"openlaneclient.CreateGroupMembershipInput": reflect.TypeOf(openlaneclient.CreateGroupMembershipInput{}),
		"openlaneclient.UpdateGroupMembershipInput": reflect.TypeOf(openlaneclient.UpdateGroupMembershipInput{}),
	}

	if err := speccli.RegisterFromFS(specFS, speccli.LoaderOptions{
		TypeResolver: speccli.StaticTypeResolver(typeMap),
		Parsers:      speccli.DefaultParsers(),
		CreateHooks:  map[string]speccli.CreateHookFactory{"groupMembershipCreate": groupMembershipCreateHook},
		UpdateHooks:  map[string]speccli.UpdateHookFactory{"groupMembershipUpdate": groupMembershipUpdateHook},
		DeleteHooks:  map[string]speccli.DeleteHookFactory{"groupMembershipDelete": groupMembershipDeleteHook},
		GetHooks:     map[string]speccli.GetHookFactory{"groupMembershipGet": groupMembershipGetHook},
	}); err != nil {
		panic(fmt.Sprintf("failed to register group members command: %v", err))
	}
}
