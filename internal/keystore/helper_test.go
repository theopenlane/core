package keystore

import "testing"

func TestHelperNaming(t *testing.T) {
	helper := NewHelper("github", "octocat")

	if helper.Name() != "Github Integration (octocat)" {
		t.Fatalf("unexpected name: %s", helper.Name())
	}
	if helper.Description() != "OAuth integration with github for octocat" {
		t.Fatalf("unexpected description: %s", helper.Description())
	}

	secret := helper.SecretName("access_token")
	if secret != "github_access_token" {
		t.Fatalf("unexpected secret name: %s", secret)
	}

	display := helper.SecretDisplayName("Github Integration", "access_token")
	if display != "Github Integration access token" {
		t.Fatalf("unexpected display name: %s", display)
	}

	desc := helper.SecretDescription("access_token")
	if desc != "Access Token for github integration" {
		t.Fatalf("unexpected secret description: %s", desc)
	}
}
