//go:build cli

package notificationpreference

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/theopenlane/core/cli/cmd"
	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/go-client/graphclient"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new notification preference",
	Run: func(cmd *cobra.Command, args []string) {
		err := create(cmd.Context())
		cobra.CheckErr(err)
	},
}

func init() {
	command.AddCommand(createCmd)

	createCmd.Flags().String("user-id", "", "user this preference applies to")
	createCmd.Flags().String("channel", "", "channel this preference applies to (e.g. EMAIL, SLACK)")
	createCmd.Flags().Bool("enabled", true, "whether this preference is enabled")
	createCmd.Flags().String("cadence", "", "delivery cadence (e.g. IMMEDIATE, DAILY, WEEKLY)")
	createCmd.Flags().String("priority", "", "optional priority override")
	createCmd.Flags().String("provider", "", "provider service for the channel")
	createCmd.Flags().String("destination", "", "destination address or endpoint for the channel")
	createCmd.Flags().StringSlice("topic-patterns", []string{}, "topic names or wildcard patterns; empty means all")
}

// createValidation validates the required fields for the command
func createValidation() (input graphclient.CreateNotificationPreferenceInput, err error) {
	userID := cmd.Config.String("user-id")
	if userID == "" {
		return input, cmd.NewRequiredFieldMissingError("user id")
	}

	input.UserID = userID

	channel := cmd.Config.String("channel")
	if channel == "" {
		return input, cmd.NewRequiredFieldMissingError("channel")
	}

	input.Channel = enums.Channel(channel)

	enabled := cmd.Config.Bool("enabled")
	input.Enabled = &enabled

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

	return input, nil
}

// create a new notification preference
func create(ctx context.Context) error {
	client, err := cmd.TokenAuth(ctx, cmd.Config)
	if err != nil || client == nil {
		client, err = cmd.SetupClientWithAuth(ctx)
		cobra.CheckErr(err)
		defer cmd.StoreSessionCookies(client)
	}

	input, err := createValidation()
	cobra.CheckErr(err)

	o, err := client.CreateNotificationPreference(ctx, input)
	cobra.CheckErr(err)

	return consoleOutput(o)
}
