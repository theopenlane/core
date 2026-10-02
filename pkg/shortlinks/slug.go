package shortlinks

import (
	"crypto/sha256"
	"encoding/base32"
	"strings"
)

// deterministicSlugLength keeps derived slugs short while leaving ~80 bits of hash
const deterministicSlugLength = 16

// slugEncoding is lowercase base32 without padding, which satisfies the service's slug rules
var slugEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// DeterministicSlug derives a stable slug from the identifying parts of a link so a retry or a
// resend of the same link reuses one record instead of minting a new one each time
func DeterministicSlug(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x1f")))

	return strings.ToLower(slugEncoding.EncodeToString(sum[:]))[:deterministicSlugLength]
}
