package githubapp

import (
	"github.com/shurcooL/githubv4"

	"github.com/theopenlane/core/internal/integrations/types"
)

var (
	DefinitionID                  = types.NewDefinitionRef("def_01K0GHAPP000000000000000001")
	GitHubClient                  = types.NewClientRef[*githubv4.Client]()
	HealthDefaultOperation        = types.NewOperationRef[HealthCheck]("health.default")
	RepositorySyncOperation       = types.NewOperationRef[RepositorySync]("repository.sync")
	VulnerabilityCollectOperation = types.NewOperationRef[VulnerabilityCollect]("vulnerability.collect")
)

const Slug = "github_app"
