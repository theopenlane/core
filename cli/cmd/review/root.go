//go:build cli

package review

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

// command represents the base review command when called without any subcommands
var command = &cobra.Command{
	Use:   "review",
	Short: "the subcommands for working with Reviews",
}

func init() {
	cmd.RootCmd.AddCommand(command)
}

// consoleOutput prints the output in the console
func consoleOutput(e any) error {
	// check if the output format is JSON and print the Reviews in JSON format
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	// check the type of the Reviews and print them in a table format
	switch v := e.(type) {
	case *graphclient.GetAllReviews:
		var nodes []*graphclient.GetAllReviews_Reviews_Edges_Node

		for _, i := range v.Reviews.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetReviews:
		var nodes []*graphclient.GetReviews_Reviews_Edges_Node

		for _, i := range v.Reviews.Edges {
			nodes = append(nodes, i.Node)
		}

		e = nodes
	case *graphclient.GetReviewByID:
		e = v.Review
	case *graphclient.CreateReview:
		e = v.CreateReview.Review
	case *graphclient.UpdateReview:
		e = v.UpdateReview.Review
	case *graphclient.DeleteReview:
		deletedTableOutput(v)
		return nil
	}

	s, err := json.Marshal(e)
	cobra.CheckErr(err)

	var list []graphclient.Review

	err = json.Unmarshal(s, &list)
	if err != nil {
		var in graphclient.Review
		err = json.Unmarshal(s, &in)
		cobra.CheckErr(err)

		list = append(list, in)
	}

	tableOutput(list)

	return nil
}

// jsonOutput prints the output in a JSON format
func jsonOutput(out any) error {
	s, err := json.Marshal(out)
	cobra.CheckErr(err)

	return cmd.JSONPrint(s)
}

// tableOutput prints the output in a table format
func tableOutput(out []graphclient.Review) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Category", "Approved", "ReviewedAt", "Reporter")
	for _, i := range out {
		category := ""
		if i.Category != nil {
			category = *i.Category
		}

		approved := false
		if i.Approved != nil {
			approved = *i.Approved
		}

		reviewedAt := ""
		if i.ReviewedAt != nil {
			reviewedAt = i.ReviewedAt.String()
		}

		reporter := ""
		if i.Reporter != nil {
			reporter = *i.Reporter
		}

		writer.AddRow(i.ID, category, approved, reviewedAt, reporter)
	}

	writer.Render()
}

// deleteTableOutput prints the deleted id in a table format
func deletedTableOutput(e *graphclient.DeleteReview) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "DeletedID")

	writer.AddRow(e.DeleteReview.DeletedID)

	writer.Render()
}
