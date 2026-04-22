package emailtemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/integrations/cli/cmd"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new email template",
	RunE: func(c *cobra.Command, _ []string) error {
		return create(c.Context())
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().StringP("key", "k", "", "stable identifier for the template (required)")
	createCmd.Flags().StringP("name", "n", "", "display name for the template (required)")
	createCmd.Flags().StringP("description", "d", "", "description of the template")
	createCmd.Flags().String("template-context", "", "runtime data context: CAMPAIGN_RECIPIENT, TRANSACTIONAL, WORKFLOW_ACTION (required)")
	createCmd.Flags().String("subject-template", "", "subject template for email notifications")
	createCmd.Flags().String("body-template", "", "body template for the email")
	createCmd.Flags().String("text-template", "", "plain text fallback template")
	createCmd.Flags().String("preheader-template", "", "preheader/preview text template")
	createCmd.Flags().String("template-format", "", "template format for rendering: TEXT, MARKDOWN, HTML")
	createCmd.Flags().String("locale", "", "locale for the template, e.g. en-US")
	createCmd.Flags().Bool("active", true, "whether the template is active")
	createCmd.Flags().String("defaults-file", "", "path to a JSON file containing template defaults (merged as base layer at render time)")
	createCmd.Flags().String("defaults-json", "", "inline JSON string containing template defaults")
	createCmd.Flags().String("email-branding-id", "", "email branding ID to attach to the template")
}

// resolveDefaults merges --defaults-file and --defaults-json into a single map.
// When both are provided, defaults-json takes precedence for overlapping keys
func resolveDefaults(filePath, inline string) (map[string]any, error) {
	out := map[string]any{}

	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrDefaultsFileUnreadable, err)
		}

		if err := json.Unmarshal(data, &out); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrDefaultsInvalid, err)
		}
	}

	if inline != "" {
		var overlay map[string]any
		if err := json.Unmarshal([]byte(inline), &overlay); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrDefaultsInvalid, err)
		}

		maps.Copy(out, overlay)
	}

	return out, nil
}

// buildCreateInput builds the CreateEmailTemplateInput from the loaded config
func buildCreateInput() (graphclient.CreateEmailTemplateInput, error) {
	var input graphclient.CreateEmailTemplateInput

	key := cmd.Config.String("key")
	if key == "" {
		return input, ErrKeyRequired
	}

	input.Key = key

	name := cmd.Config.String("name")
	if name == "" {
		return input, ErrNameRequired
	}

	input.Name = name

	templateContext := cmd.Config.String("template-context")
	if templateContext == "" {
		return input, ErrTemplateContextRequired
	}

	input.TemplateContext = enums.TemplateContext(templateContext)
	input.Description = lo.EmptyableToPtr(cmd.Config.String("description"))
	input.SubjectTemplate = lo.EmptyableToPtr(cmd.Config.String("subject-template"))
	input.BodyTemplate = lo.EmptyableToPtr(cmd.Config.String("body-template"))
	input.TextTemplate = lo.EmptyableToPtr(cmd.Config.String("text-template"))
	input.PreheaderTemplate = lo.EmptyableToPtr(cmd.Config.String("preheader-template"))
	input.Locale = lo.EmptyableToPtr(cmd.Config.String("locale"))

	if format := cmd.Config.String("template-format"); format != "" {
		f := enums.NotificationTemplateFormat(format)
		input.Format = &f
	}

	active := cmd.Config.Bool("active")
	input.Active = &active

	defaults, err := resolveDefaults(cmd.Config.String("defaults-file"), cmd.Config.String("defaults-json"))
	if err != nil {
		return input, err
	}

	if len(defaults) > 0 {
		input.Defaults = defaults
	}

	if brandingID := cmd.Config.String("email-branding-id"); brandingID != "" {
		input.EmailBrandingIDs = []string{brandingID}
	}

	return input, nil
}

// create executes the CreateEmailTemplate mutation
func create(ctx context.Context) error {
	input, err := buildCreateInput()
	if err != nil {
		return err
	}

	client, err := cmd.ConnectClient(ctx)
	if err != nil {
		return err
	}

	resp, err := client.CreateEmailTemplate(ctx, input)
	if err != nil {
		return err
	}

	tmpl := resp.CreateEmailTemplate.EmailTemplate

	headers := []string{"ID", "Key", "Name", "Format", "Active"}
	rows := [][]string{{
		tmpl.ID,
		tmpl.Key,
		tmpl.Name,
		string(tmpl.Format),
		cmd.BoolStr(tmpl.Active),
	}}

	return cmd.RenderTable(resp, headers, rows)
}
