//go:build cli

package subscribers

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/99designs/gqlgen/graphql"

	"github.com/theopenlane/core/cmd/cli/cmd"
	"github.com/theopenlane/core/cmd/cli/internal/speccli"
	"github.com/theopenlane/core/pkg/openlaneclient"
)

func buildSubscriberCreateInput() ([]*openlaneclient.CreateSubscriberInput, error) {
	emails := cmd.Config.Strings("emails")
	if len(emails) == 0 {
		return nil, cmd.NewRequiredFieldMissingError("emails")
	}

	tags := cmd.Config.Strings("tags")

	inputs := make([]*openlaneclient.CreateSubscriberInput, 0, len(emails))
	for _, email := range emails {
		address := strings.TrimSpace(email)
		if address == "" {
			return nil, cmd.NewRequiredFieldMissingError("email")
		}

		sub := &openlaneclient.CreateSubscriberInput{Email: address}
		if len(tags) > 0 {
			sub.Tags = tags
		}

		inputs = append(inputs, sub)
	}

	return inputs, nil
}

func buildCSVUpload(path string) (*graphql.Upload, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	contentType := "text/csv"
	if strings.HasSuffix(strings.ToLower(path), ".json") {
		contentType = "application/json"
	}

	return &graphql.Upload{
		File:        file,
		Filename:    info.Name(),
		Size:        info.Size(),
		ContentType: contentType,
	}, nil
}

func createSubscribersViaAPI(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	filePath := strings.TrimSpace(cmd.Config.String("file"))

	if filePath != "" {
		upload, err := buildCSVUpload(filePath)
		if err != nil {
			return nil, err
		}

		return client.CreateBulkCSVSubscriber(ctx, *upload)
	}

	inputs, err := buildSubscriberCreateInput()
	if err != nil {
		return nil, err
	}

	return client.CreateBulkSubscriber(ctx, inputs)
}

func updateSubscriberInput() (string, openlaneclient.UpdateSubscriberInput, error) {
	email := strings.TrimSpace(cmd.Config.String("email"))
	if email == "" {
		return "", openlaneclient.UpdateSubscriberInput{}, cmd.NewRequiredFieldMissingError("email")
	}

	var input openlaneclient.UpdateSubscriberInput

	if phone := strings.TrimSpace(cmd.Config.String("phone-number")); phone != "" {
		input.PhoneNumber = &phone
	}

	return email, input, nil
}

func deleteSubscriberInput() (string, *string, error) {
	email := strings.TrimSpace(cmd.Config.String("email"))
	if email == "" {
		return "", nil, cmd.NewRequiredFieldMissingError("email")
	}

	if orgID := strings.TrimSpace(cmd.Config.String("organization-id")); orgID != "" {
		return email, &orgID, nil
	}

	return email, nil, nil
}

func fetchSubscribers(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	email := strings.TrimSpace(cmd.Config.String("email"))
	if email != "" {
		return client.GetSubscriberByEmail(ctx, email)
	}

	var (
		activeFlag bool
		where      *openlaneclient.SubscriberWhereInput
	)

	if cmd.Config.Exists("active") {
		activeFlag = cmd.Config.Bool("active")
		where = &openlaneclient.SubscriberWhereInput{
			Active: &activeFlag,
		}
	}

	return client.GetSubscribers(ctx, nil, nil, nil, nil, where, nil)
}

func createSubscriber(ctx context.Context, client *openlaneclient.OpenlaneClient) (any, error) {
	filePath := strings.TrimSpace(cmd.Config.String("file"))
	if filePath == "" {
		return createSubscribersViaAPI(ctx, client)
	}

	subscribers, err := buildSubscriberCreateInput()
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(subscribers)
	if err != nil {
		return nil, err
	}

	tmpFile, err := os.CreateTemp("", "subscriber-bulk-*.json")
	if err != nil {
		return nil, err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(payload); err != nil {
		return nil, err
	}

	upload, err := speccli.UploadFromPath(tmpFile.Name())
	if err != nil {
		return nil, err
	}

	return client.CreateBulkCSVSubscriber(ctx, *upload)
}

func updateSubscriber(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.UpdateSubscriber, error) {
	email, input, err := updateSubscriberInput()
	if err != nil {
		return nil, err
	}

	return client.UpdateSubscriber(ctx, email, input)
}

func deleteSubscriber(ctx context.Context, client *openlaneclient.OpenlaneClient) (*openlaneclient.DeleteSubscriber, error) {
	email, orgID, err := deleteSubscriberInput()
	if err != nil {
		return nil, err
	}

	return client.DeleteSubscriber(ctx, email, orgID)
}
