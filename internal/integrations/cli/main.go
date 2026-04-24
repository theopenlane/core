package main

import (
	"github.com/theopenlane/core/internal/integrations/cli/cmd"
	_ "github.com/theopenlane/core/internal/integrations/cli/cmd/campaign"
	_ "github.com/theopenlane/core/internal/integrations/cli/cmd/emailtemplate"
	_ "github.com/theopenlane/core/internal/integrations/cli/cmd/emailtest"
	_ "github.com/theopenlane/core/internal/integrations/cli/cmd/integration"
	_ "github.com/theopenlane/core/internal/integrations/cli/cmd/quickstart"
)

func main() {
	cmd.Execute()
}
