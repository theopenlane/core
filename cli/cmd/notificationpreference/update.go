//go:build cli

package notificationpreference

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update an existing notification preference",
	Run: func(cmd *cobra.Command, args []string) {
		err := update(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(updateCmd)

	updateCmd.Flags().StringP("id", "i", "", "notification preference id to update")
	updateCmd.Flags().String("channel", "", "channel this preference applies to")
	updateCmd.Flags().String("enabled", "", "whether this preference is enabled (true, false)")
	updateCmd.Flags().String("cadence", "", "delivery cadence")
	updateCmd.Flags().String("priority", "", "optional priority override")
	updateCmd.Flags().String("provider", "", "provider service for the channel")
	updateCmd.Flags().String("destination", "", "destination address or endpoint")
	updateCmd.Flags().StringSlice("topic-patterns", []string{}, "topic names or wildcard patterns")
}

// updateValidation validates the required fields for the command
func updateValidation() (id string, input graphclient.UpdateNotificationPreferenceInput, err error) {
	id = cmd.Config.String("id")
	if id == "" {
		return id, input, cmd.NewRequiredFieldMissingError("notification preference id")
	}

	channel := cmd.Config.String("channel")
	if channel != "" {
		c := enums.Channel(channel)
		input.Channel = &c
	}

	enabled := cmd.Config.String("enabled")
	switch enabled {
	case "true":
		t := true
		input.Enabled = &t
	case "false":
		f := false
		input.Enabled = &f
	}

	cadence := cmd.Config.String("cadence")
	if cadence != "" {
		c := enums.NotificationCadence(cadence)
		input.Cadence = &c
	}

	priority := cmd.Config.String("priority")
	if priority != "" {
		p := enums.Priority(priority)
		input.Priority = &p
	}

	provider := cmd.Config.String("provider")
	if provider != "" {
		input.Provider = &provider
	}

	destination := cmd.Config.String("destination")
	if destination != "" {
		input.Destination = &destination
	}

	topicPatterns := cmd.Config.Strings("topic-patterns")
	if len(topicPatterns) > 0 {
		input.TopicPatterns = topicPatterns
	}

	return id, input, nil
}

// update an existing notification preference
func update(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	id, input, err := updateValidation()
	cobra.CheckErr(err)

	o, err := client.UpdateNotificationPreference(ctx, id, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
