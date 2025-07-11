package corejobs

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gertd/go-pluralize"
	"github.com/gocarina/gocsv"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"

	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	defaultHTTPTimeoutSeconds = 30
	defaultPageSize           = 2
)

var (
	errUnexpectedStatus   = fmt.Errorf("unexpected HTTP status")
	errGraphQLErrors      = fmt.Errorf("GraphQL query returned errors")
	errMissingRoot        = fmt.Errorf("missing root in response")
	errMissingEdges       = fmt.Errorf("missing edges in response")
	errMissingPageInfo    = fmt.Errorf("missing pageInfo in response")
	errMissingHasNextPage = fmt.Errorf("missing hasNextPage in pageInfo")
	errMissingEndCursor   = fmt.Errorf("missing endCursor in pageInfo")
)

// ExportContentArgs for the worker to process and update the record for the updated content
type ExportContentArgs struct {
	// ID of the export
	ExportID string `json:"export_id,omitempty"`
}

type ExportWorkerConfig struct {
	OpenlaneAPIHost  string `koanf:"openlaneAPIHost" json:"openlaneAPIHost" jsonschema:"required description=the openlane api host"`
	OpenlaneAPIToken string `koanf:"openlaneAPIToken" json:"openlaneAPIToken" jsonschema:"required description=the openlane api token"`
}

// Kind satisfies the river.Job interface
func (ExportContentArgs) Kind() string { return "export_content" }

// ExportContentWorker exports the content into csv and makes it downloadable
type ExportContentWorker struct {
	river.WorkerDefaults[ExportContentArgs]

	Config ExportWorkerConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for exporting"`

	olClient   olclient.OpenlaneClient
	httpClient *http.Client
}

// WithOpenlaneClient sets the Openlane client for the worker
// and returns the worker for method chaining
func (w *ExportContentWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *ExportContentWorker {
	w.olClient = cl
	return w
}

// WithHTTPClient sets the http client to use for the outward
// http request
func (w *ExportContentWorker) WithHTTPClient(client *http.Client) *ExportContentWorker {
	w.httpClient = client
	return w
}

// Work satisfies the river.Worker interface for the export content worker
// it creates a csv, uploads it and associates it with the export
func (w *ExportContentWorker) Work(ctx context.Context, job *river.Job[ExportContentArgs]) error {
	log.Info().Str("export_id", job.Args.ExportID).Msg("exporting content")

	if job.Args.ExportID == "" {
		return newMissingRequiredArg("export_id", ExportContentArgs{}.Kind())
	}

	if w.olClient == nil {
		cl, err := getOpenlaneClient(CustomDomainConfig{
			OpenlaneAPIHost:  w.Config.OpenlaneAPIHost,
			OpenlaneAPIToken: w.Config.OpenlaneAPIToken,
		})
		if err != nil {
			return err
		}

		w.olClient = cl
	}

	if w.httpClient == nil {
		w.httpClient = &http.Client{
			Timeout: time.Second * defaultHTTPTimeoutSeconds,
		}
	}

	export, err := w.olClient.GetExportByID(ctx, job.Args.ExportID)
	if err != nil {
		log.Error().Err(err).Str("export_id", job.Args.ExportID).Msg("failed to get export")
		return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusFailed)
	}

	ownerIDPtr := export.Export.OwnerID
	ownerID := ""
	if ownerIDPtr != nil {
		ownerID = *ownerIDPtr
	}

	exportType := strings.ToLower(export.Export.ExportType.String())
	singular := exportType
	rootQuery := pluralize.NewClient().Plural(singular)
	fields := export.Export.Fields

	if len(fields) == 0 {
		fields = []string{"id"}
	}

	query := w.buildGraphQLQuery(rootQuery, singular, fields, ownerID != "")

	var allNodes []map[string]any
	var after *string

	for {
		nodes, hasNext, nextCursor, err := w.fetchPage(ctx, query, rootQuery, ownerID, after)
		if err != nil {
			return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusFailed)
		}

		allNodes = append(allNodes, nodes...)

		if !hasNext {
			break
		}

		after = &nextCursor
	}

	if len(allNodes) == 0 {
		log.Info().Msg("no data found for export")
		return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusNodata)
	}

	csvData, err := w.marshalToCSV(allNodes)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal to CSV")
		return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusFailed)
	}

	filename := fmt.Sprintf("%s_export_%s_%s.csv", rootQuery, job.Args.ExportID, time.Now().Format("20060102_150405"))
	reader := bytes.NewReader(csvData)

	upload := &graphql.Upload{
		File:        reader,
		Filename:    filename,
		Size:        int64(len(csvData)),
		ContentType: "text/csv",
	}

	updateInput := openlaneclient.UpdateExportInput{
		Status: &enums.ExportStatusReady,
	}

	_, err = w.olClient.UpdateExport(ctx, job.Args.ExportID, updateInput, []*graphql.Upload{upload})
	if err != nil {
		log.Error().Err(err).Msg("failed to update export with file")
		return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusFailed)
	}

	return nil
}

// buildGraphQLQuery generates a query that can be used to paginate and fetch all data
//
// e.g
//
// query GetControls(
//
//	$first: Int
//	$last: Int
//	$after: Cursor
//	$before: Cursor
//	$where: ControlWhereInput
//	$orderBy: [ControlOrder!]
//
//	) {
//	  controls(
//	    first: $first
//	    last: $last
//	    after: $after
//	    before: $before
//	    where: $where
//	    orderBy: $orderBy
//	  ) {
//	    totalCount
//	    pageInfo {
//	      startCursor
//	      endCursor
//	      hasPreviousPage
//	      hasNextPage
//	    }
//	    edges {
//	      node {
//	        id
//	      }
//	    }
//	  }
//	}
func (w *ExportContentWorker) buildGraphQLQuery(root string, singular string, fields []string, hasWhere bool) string {
	if len(fields) == 0 {
		fields = []string{"id"}
	}

	fieldStr := strings.Join(fields, "\n        ")

	var varStr string
	var argStr string
	if hasWhere {
		caser := cases.Title(language.English)
		whereInputType := caser.String(singular) + "WhereInput"
		varStr = fmt.Sprintf(", $where: %s!", whereInputType)
		argStr = ", where: $where"
	}

	return fmt.Sprintf(`query ($first: Int, $after: Cursor%s) {
  %s(first: $first, after: $after%s) {
    totalCount
    pageInfo {
      hasNextPage
      endCursor
    }
    edges {
      node {
        %s
      }
    }
  }
}`, varStr, root, argStr, fieldStr)
}

func (w *ExportContentWorker) executeGraphQLQuery(ctx context.Context, query string, variables map[string]any) (map[string]any, error) {
	body := map[string]any{"query": query}
	if len(variables) > 0 {
		body["variables"] = variables
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.Config.OpenlaneAPIHost+"/query", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/graphql-response+json")
	req.Header.Set("Authorization", "Bearer "+w.Config.OpenlaneAPIToken)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", errUnexpectedStatus, resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data   map[string]any `json:"data"`
		Errors []any          `json:"errors"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("%w: %v", errGraphQLErrors, result.Errors)
	}

	return result.Data, nil
}

func (w *ExportContentWorker) marshalToCSV(nodes []map[string]any) ([]byte, error) {
	if len(nodes) == 0 {
		return nil, nil
	}

	headers := make([]string, 0)
	for k := range nodes[0] {
		headers = append(headers, k)
	}

	if len(headers) == 0 {
		return nil, nil
	}

	var buf bytes.Buffer

	wr := csv.NewWriter(&buf)

	writer := gocsv.NewSafeCSVWriter(wr)

	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	for _, node := range nodes {
		row := make([]string, len(headers))
		for i, h := range headers {
			val := node[h]
			if val == nil {
				row[i] = ""
			} else {
				row[i] = fmt.Sprint(val)
			}
		}

		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (w *ExportContentWorker) updateExportStatus(ctx context.Context, exportID string, status enums.ExportStatus) error {
	updateInput := openlaneclient.UpdateExportInput{
		Status: &status,
	}

	_, err := w.olClient.UpdateExport(ctx, exportID, updateInput, nil)
	if err != nil {
		log.Error().Err(err).
			Str("export_id", exportID).
			Str("status", string(status)).
			Msg("failed to update export status")
		return err
	}

	log.Info().
		Str("export_id", exportID).
		Str("status", string(status)).
		Msg("export status updated")

	return nil
}

func (w *ExportContentWorker) fetchPage(ctx context.Context, query, rootQuery, ownerID string, after *string) ([]map[string]any, bool, string, error) {
	vars := map[string]any{
		"first": defaultPageSize,
	}
	if after != nil {
		vars["after"] = *after
	}
	if ownerID != "" {
		vars["where"] = map[string]any{"ownerID": ownerID}
	}

	data, err := w.executeGraphQLQuery(ctx, query, vars)
	if err != nil {
		log.Error().Err(err).Msg("failed to execute GraphQL query")
		return nil, false, "", err
	}

	rootData, ok := data[rootQuery].(map[string]any)
	if !ok {
		log.Error().Msg("missing root in response")
		return nil, false, "", errMissingRoot
	}

	edges, ok := rootData["edges"].([]any)
	if !ok {
		log.Error().Msg("missing edges in response")
		return nil, false, "", errMissingEdges
	}

	nodes := make([]map[string]any, 0, len(edges))
	for _, edge := range edges {
		edgeMap, ok := edge.(map[string]any)
		if !ok {
			continue
		}
		node, ok := edgeMap["node"].(map[string]any)
		if ok {
			nodes = append(nodes, node)
		}
	}

	pageInfo, ok := rootData["pageInfo"].(map[string]any)
	if !ok {
		log.Error().Msg("missing pageInfo in response")
		return nil, false, "", errMissingPageInfo
	}

	hasNext, ok := pageInfo["hasNextPage"].(bool)
	if !ok {
		log.Error().Msg("missing hasNextPage in pageInfo")
		return nil, false, "", errMissingHasNextPage
	}

	var endCursor string
	if hasNext {
		endCursor, ok = pageInfo["endCursor"].(string)
		if !ok {
			log.Error().Msg("missing endCursor in pageInfo")
			return nil, false, "", errMissingEndCursor
		}
	}

	return nodes, hasNext, endCursor, nil
}
