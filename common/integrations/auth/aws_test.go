package auth

import (
	"testing"
	"time"

	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
)

func TestAWSMetadataFromProviderData_Defaults(t *testing.T) {
	meta, err := AWSMetadataFromProviderData(map[string]any{
		"region":          "us-east-1",
		"roleArn":         "arn:aws:iam::123:role/Test",
		"sessionDuration": "45m",
	}, "default-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.SessionName != "default-session" {
		t.Fatalf("expected default session name, got %q", meta.SessionName)
	}
	if meta.SessionDuration != 45*time.Minute {
		t.Fatalf("expected session duration, got %v", meta.SessionDuration)
	}
}

func TestAWSMetadataFromProviderData_Overrides(t *testing.T) {
	meta, err := AWSMetadataFromProviderData(map[string]any{
		"region":      "us-east-1",
		"roleArn":     "arn:aws:iam::123:role/Test",
		"sessionName": "custom",
	}, "default-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.SessionName != "custom" {
		t.Fatalf("expected custom session name, got %q", meta.SessionName)
	}
}

func TestAWSCredentialsFromPayload(t *testing.T) {
	payload := types.CredentialPayload{
		Data: models.CredentialSet{
			AccessKeyID:     "AKIA",
			SecretAccessKey: "SECRET",
			SessionToken:    "TOKEN",
			ProviderData: map[string]any{
				"accessKeyId":     "FALLBACK_AKIA",
				"secretAccessKey": "FALLBACK_SECRET",
				"sessionToken":    "FALLBACK_TOKEN",
			},
		},
	}

	creds := AWSCredentialsFromPayload(payload)
	if creds.AccessKeyID != "AKIA" {
		t.Fatalf("expected access key from credential set")
	}
	if creds.SecretAccessKey != "SECRET" {
		t.Fatalf("expected secret key from credential set")
	}
	if creds.SessionToken != "TOKEN" {
		t.Fatalf("expected session token from credential set")
	}

	payload.Data.AccessKeyID = ""
	payload.Data.SecretAccessKey = ""
	payload.Data.SessionToken = ""

	creds = AWSCredentialsFromPayload(payload)
	if creds.AccessKeyID != "FALLBACK_AKIA" {
		t.Fatalf("expected access key from provider data")
	}
	if creds.SecretAccessKey != "FALLBACK_SECRET" {
		t.Fatalf("expected secret key from provider data")
	}
	if creds.SessionToken != "FALLBACK_TOKEN" {
		t.Fatalf("expected session token from provider data")
	}
}
