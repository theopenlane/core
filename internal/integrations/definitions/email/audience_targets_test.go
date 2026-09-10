package email

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/v2/internal/audiences"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
)

func TestSetDedupe(t *testing.T) {
	set := &set{
		seen: map[string]struct{}{
			"existing@example.com": {},
		},
	}

	tt := []struct {
		name       string
		email      string
		entryAdded bool
	}{
		{
			name:  "existing email with caplocks added",
			email: "EXISTING@example.com",
		},
		{
			name:       "new email added",
			email:      "newemail@example.com",
			entryAdded: true,
		},
		{
			name:  "duplicate new email added again",
			email: "newemail@example.com",
		},
	}

	for _, v := range tt {

		recipient := audiences.ResolvedRecipient{
			AudienceMemberProjection: entityops.AudienceMemberProjection{
				Email: v.email,
			},
		}

		assert.Equal(t, v.entryAdded, set.add(recipient))
	}
}

func TestAudienceMetadata(t *testing.T) {

	metadata := getAudienceMetadata(audiences.ResolvedRecipient{
		Source:         "identity_holder",
		SourceObjectID: "idh_123",
		AudienceMemberProjection: entityops.AudienceMemberProjection{
			AudienceID: "aud_123",
			Metadata: map[string]any{
				"custom": "value",
			},
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
