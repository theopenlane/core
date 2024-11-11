package main

import (
	core "github.com/theopenlane/core/cmd/cli/cmd"

	// since the cmds are not part of the same package
	// they must all be imported in main
	_ "github.com/theopenlane/core/cmd/cli/cmd/apitokens"
	_ "github.com/theopenlane/core/cmd/cli/cmd/contact"
	_ "github.com/theopenlane/core/cmd/cli/cmd/entitlementplan"
	_ "github.com/theopenlane/core/cmd/cli/cmd/entitlementplanfeatures"
	_ "github.com/theopenlane/core/cmd/cli/cmd/entitlements"
	_ "github.com/theopenlane/core/cmd/cli/cmd/entity"
	_ "github.com/theopenlane/core/cmd/cli/cmd/entitytype"
	_ "github.com/theopenlane/core/cmd/cli/cmd/events"
	_ "github.com/theopenlane/core/cmd/cli/cmd/features"
	_ "github.com/theopenlane/core/cmd/cli/cmd/file"
	_ "github.com/theopenlane/core/cmd/cli/cmd/group"
	_ "github.com/theopenlane/core/cmd/cli/cmd/groupmembers"
	_ "github.com/theopenlane/core/cmd/cli/cmd/groupsetting"
	_ "github.com/theopenlane/core/cmd/cli/cmd/integration"
	_ "github.com/theopenlane/core/cmd/cli/cmd/invite"
	_ "github.com/theopenlane/core/cmd/cli/cmd/login"
	_ "github.com/theopenlane/core/cmd/cli/cmd/organization"
	_ "github.com/theopenlane/core/cmd/cli/cmd/organizationsetting"
	_ "github.com/theopenlane/core/cmd/cli/cmd/orgmembers"
	_ "github.com/theopenlane/core/cmd/cli/cmd/personalaccesstokens"
	_ "github.com/theopenlane/core/cmd/cli/cmd/register"
	_ "github.com/theopenlane/core/cmd/cli/cmd/reset"
	_ "github.com/theopenlane/core/cmd/cli/cmd/search"
	_ "github.com/theopenlane/core/cmd/cli/cmd/subscriber"
	_ "github.com/theopenlane/core/cmd/cli/cmd/switchcontext"
	_ "github.com/theopenlane/core/cmd/cli/cmd/task"
	_ "github.com/theopenlane/core/cmd/cli/cmd/template"
	_ "github.com/theopenlane/core/cmd/cli/cmd/user"
	_ "github.com/theopenlane/core/cmd/cli/cmd/usersetting"
	_ "github.com/theopenlane/core/cmd/cli/cmd/version"
	_ "github.com/theopenlane/core/cmd/cli/cmd/webhook"

	// history commands
	_ "github.com/theopenlane/core/cmd/cli/cmd/documentdatahistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/entitlementhistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/entityhistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/entitytypehistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/eventhistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/featurehistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/filehistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/grouphistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/groupmembershiphistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/groupsettinghistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/hushhistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/integrationhistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/oauthproviderhistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/organizationhistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/organizationsettinghistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/orgmembershiphistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/taskhistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/templatehistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/userhistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/usersettinghistory"
	_ "github.com/theopenlane/core/cmd/cli/cmd/webhookhistory"
)

func main() {
	core.Execute()
}
