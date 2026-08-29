package email

import "testing"

func TestAudienceRecipientSetDedupe(t *testing.T) {
	set := &set{
		seen: map[string]struct{}{
			"existing@example.com": {},
		},
	}

	if set.add(audienceRecipient{email: "Existing@Example.com"}) {
		t.Fatal("existing email was added")
	}

	if !set.add(audienceRecipient{email: "new@example.com"}) {
		t.Fatal("new email was not added")
	}

	if set.add(audienceRecipient{email: "NEW@example.com"}) {
		t.Fatal("duplicate normalized email was added")
	}

	if set.add(audienceRecipient{email: " "}) {
		t.Fatal("blank email was added")
	}
}

func TestAudienceTargetMetadata(t *testing.T) {
	metadata := audienceTargetMetadata(audienceRecipient{
		source:         "identity_holder",
		audienceID:     "aud_123",
		sourceObjectID: "idh_123",
		metadata: map[string]any{
			"custom": "value",
		},
	})

	if got, want := metadata[audienceTargetSourceKey], "identity_holder"; got != want {
		t.Fatalf("source metadata = %v, want %v", got, want)
	}

	if got, want := metadata[audienceTargetAudienceIDKey], "aud_123"; got != want {
		t.Fatalf("audience metadata = %v, want %v", got, want)
	}

	if got, want := metadata[audienceTargetSourceObjectKey], "idh_123"; got != want {
		t.Fatalf("source object metadata = %v, want %v", got, want)
	}

	if got, want := metadata["custom"], "value"; got != want {
		t.Fatalf("custom metadata = %v, want %v", got, want)
	}
}
