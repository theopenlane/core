//go:build cli

package campaign

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
	"github.com/theopenlane/utils/cli/tables"
)

var createWithTargetsCmd = &cobra.Command{
	Use:   "create-with-targets",
	Short: "create a new campaign with targets",
	Long: `Create a new campaign with targets in a single operation.

Targets can be specified either as a comma-separated list of emails using --emails,
or from a file using --targets-file. The file should have one target per line in the
format: email,fullname (fullname is optional).

See cli/cmd/campaign/targets_example.csv for a sample file format.

Examples:
  # Create campaign with inline emails
  openlane campaign create-with-targets --name "Q1 Review" --emails "alice@example.com,bob@example.com"

  # Create campaign with targets from file
  openlane campaign create-with-targets --name "Q1 Review" --targets-file targets.csv
`,
	Run: func(cmd *cobra.Command, args []string) {
		err := createWithTargets(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createWithTargetsCmd)

	createWithTargetsCmd.Flags().StringP("name", "n", "", "name of the campaign")
	createWithTargetsCmd.Flags().StringP("description", "d", "", "description of the campaign")
	createWithTargetsCmd.Flags().StringP("type", "y", "", "type of campaign (QUESTIONNAIRE, TRAINING, POLICY_ATTESTATION, VENDOR_ASSESSMENT, CUSTOM)")
	createWithTargetsCmd.Flags().StringP("status", "s", "", "status of the campaign")
	createWithTargetsCmd.Flags().StringP("assessment-id", "a", "", "assessment ID to associate with the campaign")
	createWithTargetsCmd.Flags().StringP("template-id", "t", "", "template ID to use for the campaign")
	createWithTargetsCmd.Flags().BoolP("is-recurring", "r", false, "whether the campaign recurs on a schedule")
	createWithTargetsCmd.Flags().StringP("recurrence-frequency", "", "", "recurrence cadence (YEARLY, QUARTERLY, BIANNUALLY, MONTHLY)")
	createWithTargetsCmd.Flags().StringP("emails", "e", "", "comma-separated list of target email addresses")
	createWithTargetsCmd.Flags().StringP("targets-file", "f", "", "path to file containing targets (email,fullname per line)")
}

// createWithTargetsValidation validates the required fields for the command
func createWithTargetsValidation() (input graphclient.CreateCampaignWithTargetsInput, err error) {
	campaignInput := &graphclient.CreateCampaignInput{}

	campaignInput.Name = cmd.Config.String("name")
	if campaignInput.Name == "" {
		return input, cmd.NewRequiredFieldMissingError("name")
	}

	description := cmd.Config.String("description")
	if description != "" {
		campaignInput.Description = &description
	}

	campaignType := cmd.Config.String("type")
	if campaignType != "" {
		campaignInput.CampaignType = enums.ToCampaignType(campaignType)
	}

	status := cmd.Config.String("status")
	if status != "" {
		campaignInput.Status = enums.ToCampaignStatus(status)
	}

	assessmentID := cmd.Config.String("assessment-id")
	if assessmentID != "" {
		campaignInput.AssessmentID = &assessmentID
	}

	templateID := cmd.Config.String("template-id")
	if templateID != "" {
		campaignInput.TemplateID = &templateID
	}

	isRecurring := cmd.Config.Bool("is-recurring")
	campaignInput.IsRecurring = &isRecurring

	recurrenceFrequency := cmd.Config.String("recurrence-frequency")
	if recurrenceFrequency != "" {
		campaignInput.RecurrenceFrequency = enums.ToFrequency(recurrenceFrequency)
	}

	input.Campaign = campaignInput

	// Parse targets from emails flag or file
	emails := cmd.Config.String("emails")
	targetsFile := cmd.Config.String("targets-file")

	if emails == "" && targetsFile == "" {
		return input, cmd.NewRequiredFieldMissingError("emails or targets-file")
	}

	var targets []*graphclient.CreateCampaignTargetInput

	if emails != "" {
		targets, err = parseEmailList(emails)
		if err != nil {
			return input, err
		}
	}

	if targetsFile != "" {
		fileTargets, err := parseTargetsFile(targetsFile)
		if err != nil {
			return input, err
		}

		targets = append(targets, fileTargets...)
	}

	if len(targets) == 0 {
		return input, cmd.NewRequiredFieldMissingError("at least one target email")
	}

	input.Targets = targets

	return input, nil
}

// parseEmailList parses a comma-separated list of emails into target inputs
func parseEmailList(emails string) ([]*graphclient.CreateCampaignTargetInput, error) {
	var targets []*graphclient.CreateCampaignTargetInput

	emailList := strings.Split(emails, ",")
	for _, email := range emailList {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}

		targets = append(targets, &graphclient.CreateCampaignTargetInput{
			Email: email,
		})
	}

	return targets, nil
}

// parseTargetsFile reads targets from a file with format: email,fullname (fullname optional)
func parseTargetsFile(filepath string) ([]*graphclient.CreateCampaignTargetInput, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []*graphclient.CreateCampaignTargetInput

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ",", 2)
		email := strings.TrimSpace(parts[0])
		if email == "" {
			continue
		}

		target := &graphclient.CreateCampaignTargetInput{
			Email: email,
		}

		if len(parts) > 1 {
			fullName := strings.TrimSpace(parts[1])
			if fullName != "" {
				target.FullName = &fullName
			}
		}

		targets = append(targets, target)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}

// createWithTargets creates a new campaign with targets
func createWithTargets(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createWithTargetsValidation()
	cobra.CheckErr(err)

	o, err := client.CreateCampaignWithTargets(ctx, input)
	cobra.CheckErr(err)

	return consoleOutputWithTargetsCreated(o)
}

// consoleOutputWithTargetsCreated prints the created campaign with its targets
func consoleOutputWithTargetsCreated(e *graphclient.CreateCampaignWithTargets) error {
	if strings.EqualFold(cmd.OutputFormat, cmd.JSONOutput) {
		return jsonOutput(e)
	}

	campaign := e.CreateCampaignWithTargets.Campaign
	targets := e.CreateCampaignWithTargets.CampaignTargets

	// Print campaign info
	fmt.Printf("Campaign created: %s (%s)\n", campaign.Name, campaign.ID)
	fmt.Printf("Targets created: %d\n\n", len(targets))

	// Print targets table
	tableOutputCreatedTargets(targets)

	return nil
}

// tableOutputCreatedTargets prints the created targets in a table format
func tableOutputCreatedTargets(targets []*graphclient.CreateCampaignWithTargets_CreateCampaignWithTargets_CampaignTargets) {
	writer := tables.NewTableWriter(command.OutOrStdout(), "ID", "Email", "FullName", "Status")
	for _, t := range targets {
		fullName := "-"
		if t.FullName != nil {
			fullName = *t.FullName
		}

		writer.AddRow(t.ID, t.Email, fullName, string(t.Status))
	}

	writer.Render()
}
