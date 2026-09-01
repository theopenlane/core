package oci

import (
	"context"

	"github.com/oracle/oci-go-sdk/v65/cloudguard"
	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/providerkit"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/logx"
)

// findingsPageSize is the number of Cloud Guard problems requested per paginated API call
const findingsPageSize = 100

// FindingsCollect collects OCI Cloud Guard problems for ingest as findings
type FindingsCollect struct{}

// IngestHandle adapts findings collection to the ingest operation registration boundary
func (f FindingsCollect) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequestConfig(cloudGuardClient, findingsSyncOperation, ErrOperationConfigInvalid, func(ctx context.Context, request types.OperationRequest, client *cloudguard.CloudGuardClient, cfg FindingsSync) ([]types.IngestPayloadSet, error) {
		return f.Run(ctx, request.Credentials, client, cfg)
	})
}

// Run collects Cloud Guard problems from the configured compartment and emits finding ingest payloads
func (FindingsCollect) Run(ctx context.Context, credentials types.CredentialBindings, c *cloudguard.CloudGuardClient, cfg FindingsSync) ([]types.IngestPayloadSet, error) {
	meta, err := resolveCredential(credentials)
	if err != nil {
		return nil, err
	}

	compartment := resolveCompartment(meta)
	if compartment == "" {
		return nil, ErrCompartmentRequired
	}

	problems, err := listProblems(ctx, c, compartment)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("compartment_id", compartment).Bool("context_cancelled", ctx.Err() != nil).Msg("oci: cloud guard problem listing failed")

		return nil, ErrListProblemsFailed
	}

	envelopes := make([]types.MappingEnvelope, 0, len(problems))

	for _, problem := range problems {
		envelope, err := providerkit.MarshalEnvelope(compartment, problemPayload(ctx, c, problem, cfg.SkipProblemDetails), ErrPayloadEncode)
		if err != nil {
			return nil, err
		}

		envelopes = append(envelopes, envelope)
	}

	return []types.IngestPayloadSet{
		{
			Schema:    entityops.SchemaFinding.Name,
			Envelopes: envelopes,
		},
	}, nil
}

// listProblems pages through every Cloud Guard problem in the compartment and its subcompartments; no time filter because resolutions and dismissals never bump timeLastDetected
func listProblems(ctx context.Context, c *cloudguard.CloudGuardClient, compartmentID string) ([]cloudguard.ProblemSummary, error) {
	problems := make([]cloudguard.ProblemSummary, 0)

	req := cloudguard.ListProblemsRequest{
		CompartmentId: lo.ToPtr(compartmentID),
		// compartments are hierarchical, so without the subtree only the root compartment reports problems
		CompartmentIdInSubtree: lo.ToPtr(true),
		AccessLevel:            cloudguard.ListProblemsAccessLevelAccessible,
		Limit:                  lo.ToPtr(findingsPageSize),
	}

	for page := 0; ; page++ {
		res, err := c.ListProblems(ctx, req)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Int("page", page).Int("collected", len(problems)).Msg("oci: cloud guard problems page request failed")

			return nil, err
		}

		problems = append(problems, res.Items...)

		if res.OpcNextPage == nil || *res.OpcNextPage == "" {
			break
		}

		req.Page = res.OpcNextPage
	}

	return problems, nil
}

// problemPayload returns the richest payload available for one problem. The list response carries no
// description or recommendation, so the full problem is fetched unless the operator opted out
func problemPayload(ctx context.Context, c *cloudguard.CloudGuardClient, problem cloudguard.ProblemSummary, skipDetails bool) any {
	if skipDetails || problem.Id == nil {
		return problem
	}

	res, err := c.GetProblem(ctx, cloudguard.GetProblemRequest{ProblemId: problem.Id})
	if err != nil {
		// one unreadable problem should not sink the whole sync, the summary still maps to a usable finding
		logx.FromContext(ctx).Warn().Err(err).Str("problem_id", lo.FromPtr(problem.Id)).Msg("oci: problem detail fetch failed, falling back to summary")

		return problem
	}

	return res.Problem
}
