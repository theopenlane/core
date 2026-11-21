package aikido

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/theopenlane/logx"
	"github.com/theopenlane/utils/mapx"
)

const (
	// pathWorkspace returns the workspace the credentials belong to
	pathWorkspace = pathAPI + "/workspace"
	// pathIssuesExport exports issues across the workspace
	pathIssuesExport = pathAPI + "/issues/export"
	// pathCodeRepositories lists the code repositories Aikido scans
	pathCodeRepositories = pathAPI + "/repositories/code"
	// pathContainers lists the container images Aikido scans
	pathContainers = pathAPI + "/containers"
	// pathClouds lists the connected cloud environments
	pathClouds = pathAPI + "/clouds"
	// pathDomains lists the domains Aikido monitors
	pathDomains = pathAPI + "/domains"
	// pathIssueGroups lists issue groups, the endpoint name says open but filter_status selects any status
	pathIssueGroups = pathAPI + "/open-issue-groups"
	// pathIssueDetailsBulk returns issue details the export lacks, such as ignore reasons and reachability
	pathIssueDetailsBulk = pathAPI + "/issues/detail/bulk"
	// pathAutofixHistory lists the autofix tasks Aikido ran
	pathAutofixHistory = pathAPI + "/autofix/history"
	// pathVirtualMachines lists the cloud virtual machines Aikido scans
	pathVirtualMachines = pathAPI + "/virtual-machines"
	// pathDevices lists the endpoint protection devices, unpaged
	pathDevices = pathAPI + "/endpoint-protection/devices"
	// pathSubdomainsFormat lists the subdomains discovered under one monitored domain
	pathSubdomainsFormat = pathDomains + "/%s/subdomains"
	// pathUsers lists the workspace users, unpaged
	pathUsers = pathAPI + "/users"
	// pathTeams lists the workspace teams
	pathTeams = pathAPI + "/teams"
	// pathComplianceOverviewFormat returns the compliance report of one framework
	pathComplianceOverviewFormat = pathAPI + "/report/%s/overview"
	// pathPentests lists the AI pentest assessments
	pathPentests = pathAPI + "/pentests/assessments"
	// pathCodeAudits lists the AI code analysis assessments
	pathCodeAudits = pathAPI + "/pentests/code-audits"
	// pathAttackAnalysisFormat returns the attack analysis of one AI pentest issue
	pathAttackAnalysisFormat = pathAPI + "/pentests/issues/%s/attackAnalysis"
	// pathCVEFormat returns the Aikido Intel details of one CVE or AIKIDO id
	pathCVEFormat = pathAPI + "/cve/%s"
)

// headerHasNextPage reports whether another page follows on endpoints that send it
const headerHasNextPage = "X-Has-Next-Page"

// issueGroupStatuses are swept one at a time since the issue group listing filters on a single status
var issueGroupStatuses = []string{"open", "closed", "snoozed", "ignored"}

const (
	// issueStatusAll requests open, ignored, snoozed and closed issues so a resolution in Aikido closes the record here
	issueStatusAll = "all"
	// issueExportPageSize is the export page size, the documented maximum keeps calls low under the 20 per minute limit
	issueExportPageSize = 5000
	// assetPageSize is the page size for the asset inventory listings, the documented maximum for most of them
	assetPageSize = 100
	// codeRepositoryPageSize is the documented maximum page size for code repositories
	codeRepositoryPageSize = 200
	// issueGroupPageSize is the documented maximum page size for issue groups
	issueGroupPageSize = 100
	// issueDetailBatchSize is the documented maximum number of issue ids per bulk detail call
	issueDetailBatchSize = 100
	// autofixPageSize is the documented maximum page size for the autofix history
	autofixPageSize = 20
	// virtualMachinePageSize is the documented default page size for virtual machines, which states no maximum
	virtualMachinePageSize = 20
	// teamPageSize is the documented maximum page size for teams
	teamPageSize = 100
	// assessmentPageSize is the documented maximum page size for pentest and code analysis assessments
	assessmentPageSize = 100
	// maxPages bounds a single listing so a misbehaving API can never page forever
	maxPages = 200
)

// workspaceResponse holds the workspace details reported for the authenticated credentials
type workspaceResponse struct {
	// ID is the numeric Aikido workspace identifier
	ID int64 `json:"id"`
	// Name is the workspace name
	Name string `json:"name"`
	// LinkedProvider is the source control provider linked to the workspace
	LinkedProvider string `json:"linked_provider"`
}

// identifier renders the workspace ID as the stable string used for installation identity
func (w workspaceResponse) identifier() string {
	if w.ID == 0 {
		return ""
	}

	return strconv.FormatInt(w.ID, 10)
}

// workspace returns the workspace details for the authenticated credentials
func (c *APIClient) workspace(ctx context.Context) (workspaceResponse, error) {
	workspace := workspaceResponse{}
	if _, err := c.get(ctx, pathWorkspace, nil, &workspace); err != nil {
		return workspaceResponse{}, err
	}

	return workspace, nil
}

// listPage fetches one zero-based page of a list endpoint and reports whether another page follows,
// trusting the X-Has-Next-Page header when sent since filtered pages can be short with more remaining
func (c *APIClient) listPage(ctx context.Context, path string, params map[string]string, page, perPage int) ([]map[string]any, bool, error) {
	query := maps.Clone(params)
	if query == nil {
		query = map[string]string{}
	}

	query["page"] = strconv.Itoa(page)
	query["per_page"] = strconv.Itoa(perPage)

	var records []map[string]any

	header, err := c.get(ctx, path, query, &records)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msgf("aikido: error getting object %s", path)

		// a missing scope stays visible so the installation can say which permission to grant
		if errors.Is(err, ErrUnauthorized) {
			return nil, false, ErrUnauthorized
		}

		return nil, false, ErrAPIRequest
	}

	if next := header.Get(headerHasNextPage); next != "" {
		return records, next == "true", nil
	}

	return records, len(records) >= perPage, nil
}

// listAll pages through a list endpoint until the last page, or a page that adds no unseen record
// in case the API ignores the page parameter and replays results
func (c *APIClient) listAll(ctx context.Context, path string, params map[string]string, perPage int) ([]map[string]any, error) {
	var (
		records []map[string]any
		seen    = map[string]struct{}{}
	)

	for page := range maxPages {
		if err := ctx.Err(); err != nil {
			logx.FromContext(ctx).Error().Err(err).Msgf("aikido: error retrieving objects from %s", path)
			return nil, ErrAPIRequest
		}

		batch, more, err := c.listPage(ctx, path, params, page, perPage)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msgf("aikido: error retrieving objects from %s", path)
			return nil, err
		}

		added := 0

		for _, record := range batch {
			// ids are numeric on most listings and uuid strings on others, such as subdomains
			id := fmt.Sprint(record["id"])
			if _, duplicate := seen[id]; duplicate {
				continue
			}

			seen[id] = struct{}{}
			records = append(records, record)
			added++
		}

		if !more || added == 0 {
			return records, nil
		}
	}

	return nil, ErrPageLimit
}

// exportIssues exports every issue in the workspace regardless of status
func (c *APIClient) exportIssues(ctx context.Context) ([]map[string]any, error) {
	return c.listAll(ctx, pathIssuesExport, map[string]string{"format": "json", "filter_status": issueStatusAll}, issueExportPageSize)
}

// issueGroups lists the issue groups in every status keyed by group id, groups carry the description and remediation issues lack
func (c *APIClient) issueGroups(ctx context.Context) (map[float64]map[string]any, error) {
	groups := map[float64]map[string]any{}

	for _, status := range issueGroupStatuses {
		records, err := c.listAll(ctx, pathIssueGroups, map[string]string{"filter_status": status}, issueGroupPageSize)
		if err != nil {
			return nil, err
		}

		indexByID(groups, records)
	}

	return groups, nil
}

// indexByID adds each record to index under its numeric id
func indexByID(index map[float64]map[string]any, records []map[string]any) {
	for _, record := range records {
		if id := mapx.Value[float64](record, "id"); id != 0 {
			index[id] = record
		}
	}
}

// issueDetails fetches the bulk issue details keyed by issue id, batching ids at the documented maximum
func (c *APIClient) issueDetails(ctx context.Context, ids []string) (map[float64]map[string]any, error) {
	details := make(map[float64]map[string]any, len(ids))

	for batch := range slices.Chunk(ids, issueDetailBatchSize) {
		var records []map[string]any

		if _, err := c.get(ctx, pathIssueDetailsBulk, map[string]string{"issue_ids": strings.Join(batch, ",")}, &records); err != nil {
			return nil, err
		}

		indexByID(details, records)
	}

	return details, nil
}

// dataleaksResponse wraps the leaked credentials Aikido found for one domain
type dataleaksResponse struct {
	// Dataleaks are the leaked credentials found for the domain
	Dataleaks []map[string]any `json:"dataleaks"`
}

// domainDataleaks lists the leaked credentials Aikido found for one monitored domain
func (c *APIClient) domainDataleaks(ctx context.Context, domainID string) ([]map[string]any, error) {
	var response dataleaksResponse

	if _, err := c.get(ctx, pathDomains+"/"+domainID+"/dataleaks", nil, &response); err != nil {
		return nil, err
	}

	return response.Dataleaks, nil
}

// autofixIssuesResponse wraps the issues one autofix task targets
type autofixIssuesResponse struct {
	// TargetedIssues are the issues the autofix task addresses
	TargetedIssues []map[string]any `json:"targeted_issues"`
}

// autofixIssueIDs returns the ids of the issues one autofix task targets
func (c *APIClient) autofixIssueIDs(ctx context.Context, taskID string) ([]string, error) {
	var response autofixIssuesResponse

	if _, err := c.get(ctx, pathAutofixHistory+"/"+taskID+"/issues", nil, &response); err != nil {
		return nil, err
	}

	return issueIDs(response.TargetedIssues), nil
}

// users lists the workspace users, including inactive ones so a deactivation in Aikido reaches the directory
func (c *APIClient) users(ctx context.Context, params map[string]string) ([]map[string]any, error) {
	query := maps.Clone(params)
	if query == nil {
		query = map[string]string{}
	}

	query["include_inactive"] = "1"

	var records []map[string]any

	if _, err := c.get(ctx, pathUsers, query, &records); err != nil {
		return nil, err
	}

	return records, nil
}

// fetchEach fetches the record at pathFormat for each key, keyed by key, skipping a key that fails and stopping
// when the scope is missing or Aikido keeps rate limiting, since every remaining call would fail the same way
func (c *APIClient) fetchEach(ctx context.Context, keys []string, pathFormat string) (map[string]map[string]any, error) {
	records := make(map[string]map[string]any, len(keys))

	for _, key := range keys {
		var record map[string]any

		if _, err := c.get(ctx, fmt.Sprintf(pathFormat, key), nil, &record); err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			if errors.Is(err, ErrUnauthorized) || errors.Is(err, ErrRateLimited) {
				logx.FromContext(ctx).Warn().Err(err).Str("path", pathFormat).Int("fetched", len(records)).Int("requested", len(keys)).Msg("aikido: optional detail unavailable, ingesting without the rest")

				return records, nil
			}

			logx.FromContext(ctx).Warn().Err(err).Str("path", pathFormat).Str("key", key).Msg("aikido: optional detail unavailable, skipping it")

			continue
		}

		records[key] = record
	}

	return records, nil
}
