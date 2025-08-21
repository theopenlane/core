package corejobs

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gertd/go-pluralize"
	"github.com/gocarina/gocsv"
	"github.com/riverqueue/river"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/httpsling"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/theopenlane/core/pkg/corejobs/internal/olclient"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/openlaneclient"
	"github.com/theopenlane/iam/auth"
)

const (
	defaultHTTPTimeoutSeconds = 30
	defaultPageSize           = 2
)

var (
	// ErrUnexpectedStatus is returned when an HTTP request returns a status code other than 200
	ErrUnexpectedStatus = errors.New("unexpected HTTP status")
	// ErrGraphQLMessage is returned when an error message exists in the response
	ErrGraphQLMessage = errors.New("GraphQL error")
	// ErrUnknownGraphQLError is returned when an GraphQL error occurs but no specific message is available
	ErrUnknownGraphQLError = errors.New("an unknown error occurred")
	// ErrMissingRoot is returned when the GraphQL response is missing the expected root field
	ErrMissingRoot = errors.New("missing root in response")
	// ErrMissingEdges is returned when the response is missing the edges field expected for the data
	ErrMissingEdges = errors.New("missing edges in response")
	// ErrMissingPageInfo is returned when the response is missing pagination data
	ErrMissingPageInfo = errors.New("missing pageInfo in response")
	// ErrMissingHasNextPage is returned when pagination data is missing the hasNextPage field
	ErrMissingHasNextPage = errors.New("missing hasNextPage in pageInfo")
	// ErrMissingEndCursor is returned when pagination data is missing the endCursor field needed for pagination
	ErrMissingEndCursor = errors.New("missing endCursor in pageInfo")
)

// ExportContentArgs for the worker to process and update the record for the updated content
type ExportContentArgs struct {
	// ID of the export
	ExportID string `json:"export_id,omitempty"`
	// UserID of the user who requested the export (for system admin context)
	UserID string `json:"user_id,omitempty"`
	// OrganizationID of the organization context for the export
	OrganizationID string `json:"organization_id,omitempty"`
}

type ExportWorkerConfig struct {
	Enabled bool `koanf:"enabled" json:"enabled" jsonschema:"required description=whether the export worker is enabled"`

	OpenlaneAPIHost  string `koanf:"openlaneAPIHost" json:"openlaneAPIHost" jsonschema:"required description=the openlane api host"`
	OpenlaneAPIToken string `koanf:"openlaneAPIToken" json:"openlaneAPIToken" jsonschema:"required description=the openlane api token"`
}

// Kind satisfies the river.Job interface
func (ExportContentArgs) Kind() string { return "export_content" }

// ExportContentWorker exports the content into csv and makes it downloadable
type ExportContentWorker struct {
	river.WorkerDefaults[ExportContentArgs]

	Config ExportWorkerConfig `koanf:"config" json:"config" jsonschema:"description=the configuration for exporting"`

	olClient  olclient.OpenlaneClient
	requester *httpsling.Requester
}

// WithOpenlaneClient sets the Openlane client for the worker
// and returns the worker for method chaining
func (w *ExportContentWorker) WithOpenlaneClient(cl olclient.OpenlaneClient) *ExportContentWorker {
	w.olClient = cl
	return w
}

// WithRequester sets the httpsling requester to use for HTTP requests
func (w *ExportContentWorker) WithRequester(requester *httpsling.Requester) *ExportContentWorker {
	w.requester = requester
	return w
}

// Work satisfies the river.Worker interface for the export content worker
// it creates a csv, uploads it and associates it with the export
func (w *ExportContentWorker) Work(ctx context.Context, job *river.Job[ExportContentArgs]) error {
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

	if w.requester == nil {
		var err error
		w.requester, err = httpsling.New(
			httpsling.URL(w.Config.OpenlaneAPIHost),
			httpsling.BearerAuth(w.Config.OpenlaneAPIToken),
			httpsling.Header("Content-Type", "application/json"),
			httpsling.Header("Accept", "application/graphql-response+json"),
		)
		if err != nil {
			return err
		}
	}

	export, err := w.olClient.GetExportByID(ctx, job.Args.ExportID)
	if err != nil {
		log.Error().Err(err).Str("export_id", job.Args.ExportID).Msg("failed to get export")
		return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusFailed, err)
	}

	ownerIDPtr := export.Export.OwnerID
	ownerID := ""
	if ownerIDPtr != nil {
		ownerID = *ownerIDPtr
	}

	var filterMap map[string]any
	filtersPtr := export.Export.Filters
	if filtersPtr != nil {
		filters := *filtersPtr
		if filters != "" {
			if err := json.Unmarshal([]byte(filters), &filterMap); err != nil {
				log.Error().Err(err).Msg("failed to parse filters")
				return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusFailed, err)
			}

		}
	}

	where := make(map[string]any)

	if _, exists := filterMap["ownerID"]; !exists && ownerID != "" {
		where["ownerID"] = ownerID
	}

	maps.Copy(where, filterMap)

	hasWhere := len(where) > 0

	exportType := strings.ToLower(export.Export.ExportType.String())
	singular := exportType
	rootQuery := pluralize.NewClient().Plural(singular)

	fields := export.Export.Fields

	query := w.buildGraphQLQuery(rootQuery, singular, fields, hasWhere)

	var allNodes []map[string]any
	var after *string

	for {
		nodes, hasNext, nextCursor, err := w.fetchPage(ctx, query, rootQuery, after, where, job.Args)
		if err != nil {
			return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusFailed, err)
		}

		allNodes = append(allNodes, nodes...)

		if !hasNext {
			break
		}

		after = &nextCursor
	}

	if len(allNodes) == 0 {
		log.Info().Msg("no data found for export")
		return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusNodata, nil)
	}

	csvData, err := w.marshalToCSV(allNodes)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal to CSV")
		return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusFailed, err)
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

	_, err = w.olClient.UpdateExport(ctx, job.Args.ExportID, updateInput, []*graphql.Upload{upload}, openlaneclient.WithImpersonationInterceptor(job.Args.UserID, job.Args.OrganizationID))
	if err != nil {
		log.Error().Err(err).Msg("failed to update export with file")
		return w.updateExportStatus(ctx, job.Args.ExportID, enums.ExportStatusFailed, err)
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

// extractErrors converts a slice of errors from the request into one
func extractErrors(errs []any) error {
	if len(errs) == 0 {
		return nil
	}

	var errMsgs []error
	for _, e := range errs {
		if msg, ok := e.(map[string]any); ok {
			if m, ok := msg["message"].(string); ok {
				errMsgs = append(errMsgs, fmt.Errorf("%w: %s", ErrGraphQLMessage, m))
			}
		}
	}

	if len(errMsgs) > 0 {
		return errors.Join(errMsgs...)
	}

	return ErrUnknownGraphQLError
}

func (w *ExportContentWorker) executeGraphQLQuery(ctx context.Context, query string, variables map[string]any, jobArgs ExportContentArgs) (map[string]any, error) {
	body := map[string]any{"query": query}
	if len(variables) > 0 {
		body["variables"] = variables
	}

	// Prepare request options
	opts := []httpsling.Option{
		httpsling.Post("/query"),
		httpsling.Body(body),
	}

	// Add user context headers if provided (for system admin operations)
	if jobArgs.UserID != "" {
		opts = append(opts, httpsling.Header(auth.UserIDHeader, jobArgs.UserID))
	}
	if jobArgs.OrganizationID != "" {
		opts = append(opts, httpsling.Header(auth.OrganizationIDHeader, jobArgs.OrganizationID))
	}

	var result struct {
		Data   map[string]any `json:"data"`
		Errors []any          `json:"errors"`
	}

	resp, err := w.requester.ReceiveWithContext(ctx, &result, opts...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if len(result.Errors) > 0 {
		return nil, extractErrors(result.Errors)
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

func (w *ExportContentWorker) updateExportStatus(ctx context.Context, exportID string, status enums.ExportStatus, err error) error {
	updateInput := openlaneclient.UpdateExportInput{
		Status: &status,
	}

	if status == enums.ExportStatusFailed && err != nil {
		msg := err.Error()
		updateInput.ErrorMessage = &msg
	}

	_, err = w.olClient.UpdateExport(ctx, exportID, updateInput, nil)
	if err != nil {
		log.Error().Err(err).
			Str("export_id", exportID).
			Str("status", string(status)).
			Msg("failed to update export status")
		return err
	}

	log.Info().Str("export_id", exportID).Msg("export status updated")

	return nil
}

func (w *ExportContentWorker) fetchPage(ctx context.Context, query, rootQuery string, after *string, where map[string]any, jobArgs ExportContentArgs) ([]map[string]any, bool, string, error) {
	vars := map[string]any{"first": defaultPageSize}
	if after != nil {
		vars["after"] = *after
	}
	if len(where) > 0 {
		vars["where"] = where
	}

	data, err := w.executeGraphQLQuery(ctx, query, vars, jobArgs)
	if err != nil {
		log.Error().Err(err).Msg("failed to execute GraphQL query")
		return nil, false, "", err
	}

	rootData, ok := data[rootQuery].(map[string]any)
	if !ok {
		log.Error().Msg("missing root in response")
		return nil, false, "", ErrMissingRoot
	}

	edges, ok := rootData["edges"].([]any)
	if !ok {
		log.Error().Msg("missing edges in response")
		return nil, false, "", ErrMissingEdges
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
		return nil, false, "", ErrMissingPageInfo
	}

	hasNext, ok := pageInfo["hasNextPage"].(bool)
	if !ok {
		log.Error().Msg("missing hasNextPage in pageInfo")
		return nil, false, "", ErrMissingHasNextPage
	}

	var endCursor string
	if hasNext {
		endCursor, ok = pageInfo["endCursor"].(string)
		if !ok {
			log.Error().Msg("missing endCursor in pageInfo")
			return nil, false, "", ErrMissingEndCursor
		}
	}

	return nodes, hasNext, endCursor, nil
}
