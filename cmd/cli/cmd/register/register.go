//go:build cli

package register

import (
	"context"
	"encoding/json"

	"github.com/theopenlane/core/cmd/cli/cmd"
	models "github.com/theopenlane/core/pkg/openapi"
)

func buildRegisterRequest() (*models.RegisterRequest, error) {
	email := cmd.Config.String("email")
	if email == "" {
		return nil, cmd.NewRequiredFieldMissingError("email")
	}

	firstName := cmd.Config.String("first-name")
	if firstName == "" {
		return nil, cmd.NewRequiredFieldMissingError("first name")
	}

	lastName := cmd.Config.String("last-name")
	if lastName == "" {
		return nil, cmd.NewRequiredFieldMissingError("last name")
	}

	password := cmd.Config.String("password")
	if password == "" {
		return nil, cmd.NewRequiredFieldMissingError("password")
	}

	return &models.RegisterRequest{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  password,
	}, nil
}

func executeRegister(ctx context.Context) ([]byte, error) {
	client, err := cmd.SetupClient(ctx)
	if err != nil {
		return nil, err
	}

	input, err := buildRegisterRequest()
	if err != nil {
		return nil, err
	}

	registration, err := client.Register(ctx, input)
	if err != nil {
		return nil, err
	}

	return json.Marshal(registration)
}
