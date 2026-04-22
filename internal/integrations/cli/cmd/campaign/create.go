package campaign

import (
	"bufio"
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

// targetCommentPrefix identifies comment lines in a targets file
const targetCommentPrefix = "#"

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new campaign with one or more targets",
	Long: `Create a campaign and attach its initial targets in a single operation.

Targets may be supplied inline via --emails (comma-separated) and/or loaded
from --targets-file (one target per line as "email,fullname"; fullname is
optional; lines starting with # are ignored).`,
	RunE: func(c *cobra.Command, _ []string) error {
		return create(c.Context())
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("name", "n", "", "name of the campaign (required)")
	createCmd.Flags().StringP("description", "d", "", "description of the campaign")
	createCmd.Flags().StringP("type", "y", "", "campaign type: QUESTIONNAIRE, TRAINING, POLICY_ATTESTATION, VENDOR_ASSESSMENT, CUSTOM")
	createCmd.Flags().StringP("status", "s", "", "campaign status")
	createCmd.Flags().StringP("assessment-id", "a", "", "assessment ID to associate with the campaign")
	createCmd.Flags().StringP("template-id", "t", "", "notification template ID for the campaign")
	createCmd.Flags().String("email-template-id", "", "email template ID (catalog entry) driving campaign sends")
	createCmd.Flags().BoolP("is-recurring", "r", false, "whether the campaign recurs on a schedule")
	createCmd.Flags().String("recurrence-frequency", "", "recurrence cadence: YEARLY, QUARTERLY, BIANNUALLY, MONTHLY")
	createCmd.Flags().StringP("emails", "e", "", "comma-separated list of target email addresses")
	createCmd.Flags().StringP("targets-file", "f", "", "path to file containing targets (email,fullname per line)")
}

// buildCreateInput builds the CreateCampaignWithTargetsInput from config + flags
func buildCreateInput() (graphclient.CreateCampaignWithTargetsInput, error) {
	var input graphclient.CreateCampaignWithTargetsInput

	name := cmd.Config.String("name")
	if name == "" {
		return input, ErrNameRequired
	}

	campaign := &graphclient.CreateCampaignInput{Name: name}
	campaign.Description = lo.EmptyableToPtr(cmd.Config.String("description"))
	campaign.AssessmentID = lo.EmptyableToPtr(cmd.Config.String("assessment-id"))
	campaign.TemplateID = lo.EmptyableToPtr(cmd.Config.String("template-id"))
	campaign.EmailTemplateID = lo.EmptyableToPtr(cmd.Config.String("email-template-id"))

	if t := cmd.Config.String("type"); t != "" {
		campaign.CampaignType = enums.ToCampaignType(t)
	}

	if s := cmd.Config.String("status"); s != "" {
		campaign.Status = enums.ToCampaignStatus(s)
	}

	if f := cmd.Config.String("recurrence-frequency"); f != "" {
		campaign.RecurrenceFrequency = enums.ToFrequency(f)
	}

	isRecurring := cmd.Config.Bool("is-recurring")
	campaign.IsRecurring = &isRecurring

	input.Campaign = campaign

	targets, err := collectTargets(cmd.Config.String("emails"), cmd.Config.String("targets-file"))
	if err != nil {
		return input, err
	}

	if len(targets) == 0 {
		return input, ErrTargetsRequired
	}

	input.Targets = targets

	return input, nil
}

// collectTargets resolves inline + file-based targets into a single slice
func collectTargets(inline, filePath string) ([]*graphclient.CreateCampaignTargetInput, error) {
	var targets []*graphclient.CreateCampaignTargetInput

	if inline != "" {
		targets = append(targets, parseInlineEmails(inline)...)
	}

	if filePath != "" {
		fileTargets, err := parseTargetsFile(filePath)
		if err != nil {
			return nil, err
		}

		targets = append(targets, fileTargets...)
	}

	return targets, nil
}

// parseInlineEmails parses a comma-separated list of email addresses
func parseInlineEmails(raw string) []*graphclient.CreateCampaignTargetInput {
	parts := strings.Split(raw, ",")
	targets := make([]*graphclient.CreateCampaignTargetInput, 0, len(parts))

	for _, p := range parts {
		email := strings.TrimSpace(p)
		if email == "" {
			continue
		}

		targets = append(targets, &graphclient.CreateCampaignTargetInput{Email: email})
	}

	return targets
}

// parseTargetsFile reads targets from a file, one per line, email[,fullname]
func parseTargetsFile(path string) ([]*graphclient.CreateCampaignTargetInput, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	var targets []*graphclient.CreateCampaignTargetInput

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, targetCommentPrefix) {
			continue
		}

		parts := strings.SplitN(line, ",", 2)

		email := strings.TrimSpace(parts[0])
		if email == "" {
			continue
		}

		target := &graphclient.CreateCampaignTargetInput{Email: email}
		if len(parts) > 1 {
			target.FullName = lo.EmptyableToPtr(strings.TrimSpace(parts[1]))
		}

		targets = append(targets, target)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return targets, nil
}

// create executes the CreateCampaignWithTargets mutation and prints a summary
func create(ctx context.Context) error {
	input, err := buildCreateInput()
	if err != nil {
		return err
	}

	client, err := cmd.ConnectClient(ctx)
	if err != nil {
		return err
	}

	resp, err := client.CreateCampaignWithTargets(ctx, input)
	if err != nil {
		return err
	}

	c := resp.CreateCampaignWithTargets.Campaign

	headers := []string{"CampaignID", "Name", "Type", "Status", "Targets"}
	rows := [][]string{{
		c.ID,
		c.Name,
		string(c.CampaignType),
		string(c.Status),
		strconv.Itoa(len(resp.CreateCampaignWithTargets.CampaignTargets)),
	}}

	return cmd.RenderTable(resp, headers, rows)
}
