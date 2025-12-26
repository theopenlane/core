//go:build cli

package main

import (
	core "github.com/theopenlane/cli/cmd"

	// since the cmds are not part of the same package
	// they must all be imported in main
	_ "github.com/theopenlane/cli/cmd/apitokens"
	_ "github.com/theopenlane/cli/cmd/contact"
	_ "github.com/theopenlane/cli/cmd/control"
	_ "github.com/theopenlane/cli/cmd/controlimplementation"
	_ "github.com/theopenlane/cli/cmd/controlobjective"
	_ "github.com/theopenlane/cli/cmd/customdomain"
	_ "github.com/theopenlane/cli/cmd/dnsverification"
	_ "github.com/theopenlane/cli/cmd/documentdata"
	_ "github.com/theopenlane/cli/cmd/entity"
	_ "github.com/theopenlane/cli/cmd/entitytype"
	_ "github.com/theopenlane/cli/cmd/evidence"
	_ "github.com/theopenlane/cli/cmd/file"
	_ "github.com/theopenlane/cli/cmd/group"
	_ "github.com/theopenlane/cli/cmd/groupmembers"
	_ "github.com/theopenlane/cli/cmd/groupsetting"
	_ "github.com/theopenlane/cli/cmd/integration"
	_ "github.com/theopenlane/cli/cmd/internalpolicy"
	_ "github.com/theopenlane/cli/cmd/invite"
	_ "github.com/theopenlane/cli/cmd/jobresult"
	_ "github.com/theopenlane/cli/cmd/jobrunnertoken"
	_ "github.com/theopenlane/cli/cmd/jobtemplate"
	_ "github.com/theopenlane/cli/cmd/login"
	_ "github.com/theopenlane/cli/cmd/mappabledomain"
	_ "github.com/theopenlane/cli/cmd/mappedcontrol"
	_ "github.com/theopenlane/cli/cmd/narrative"
	_ "github.com/theopenlane/cli/cmd/organization"
	_ "github.com/theopenlane/cli/cmd/organizationsetting"
	_ "github.com/theopenlane/cli/cmd/orgmembers"
	_ "github.com/theopenlane/cli/cmd/orgsubscription"
	_ "github.com/theopenlane/cli/cmd/personalaccesstokens"
	_ "github.com/theopenlane/cli/cmd/procedure"
	_ "github.com/theopenlane/cli/cmd/program"
	_ "github.com/theopenlane/cli/cmd/programmembers"
	_ "github.com/theopenlane/cli/cmd/register"
	_ "github.com/theopenlane/cli/cmd/reset"
	_ "github.com/theopenlane/cli/cmd/risk"
	_ "github.com/theopenlane/cli/cmd/scheduledjob"
	_ "github.com/theopenlane/cli/cmd/search"
	_ "github.com/theopenlane/cli/cmd/standard"
	_ "github.com/theopenlane/cli/cmd/subcontrol"
	_ "github.com/theopenlane/cli/cmd/subprocessor"
	_ "github.com/theopenlane/cli/cmd/subscriber"
	_ "github.com/theopenlane/cli/cmd/switchcontext"
	_ "github.com/theopenlane/cli/cmd/task"
	_ "github.com/theopenlane/cli/cmd/template"
	_ "github.com/theopenlane/cli/cmd/trustcenter"
	_ "github.com/theopenlane/cli/cmd/trustcentercompliance"
	_ "github.com/theopenlane/cli/cmd/trustcenterdoc"
	_ "github.com/theopenlane/cli/cmd/trustcenterdomain"
	_ "github.com/theopenlane/cli/cmd/trustcenternda"
	_ "github.com/theopenlane/cli/cmd/trustcentersubprocessors"
	_ "github.com/theopenlane/cli/cmd/trustcenterwatermarkconfig"
	_ "github.com/theopenlane/cli/cmd/user"
	_ "github.com/theopenlane/cli/cmd/usersetting"
	_ "github.com/theopenlane/cli/cmd/version"
)

func main() {
	core.Execute()
}
