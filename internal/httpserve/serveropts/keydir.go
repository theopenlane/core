package serveropts

import (
	"context"
	"crypto"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/oklog/ulid/v2"
	"github.com/samber/lo"
	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/v2/pkg/logx"
)

const (
	pemExtension  = ".pem"
	certExtension = ".crt"

	// signHandoffSlack pads the no-sign window to absorb clock skew and renewal jitter
	signHandoffSlack = time.Hour

	// minCertAge lets a new key's JWKS entry propagate to validator caches before it signs
	minCertAge = time.Hour
)

// keyDirEntry is a signing key discovered in the key directory
type keyDirEntry struct {
	kid       string
	path      string
	signer    crypto.Signer
	notBefore time.Time
	notAfter  time.Time
	hasCert   bool
	// legacyKID is the filename-derived kid keys installed before thumbprint kids were
	// signed under; outstanding tokens carry it, so it keeps verifying until they age
	// out. The field and every branch on it are dead once no ULID-named key files remain
	// mounted, and can be deleted with them
	legacyKID string
}

// headroom is how long the entry's certificate remains valid
func (e keyDirEntry) headroom(now time.Time) time.Duration {
	return e.notAfter.Sub(now)
}

// signingMargin is the headroom a key needs for every token it can sign to expire while
// the key is still valid
func signingMargin(conf tokens.Config) time.Duration {
	return max(
		conf.AccessDuration,
		conf.RefreshDuration,
		conf.AssessmentAccessDuration,
		conf.TrustCenterNDARequestAccessDuration,
	) + signHandoffSlack
}

// tokenConfigFromKeyDir returns a copy of base with the keys in dir merged in, used to
// bootstrap the token manager before the archive is reachable
func tokenConfigFromKeyDir(base tokens.Config, dir string) (tokens.Config, error) {
	entries, err := scanKeyDir(dir)
	if err != nil {
		return tokens.Config{}, err
	}

	conf := base
	conf.Keys = make(map[string]string, len(entries)+len(base.Keys))

	for _, entry := range entries {
		conf.Keys[entry.kid] = entry.path

		if entry.legacyKID != "" {
			conf.Keys[entry.legacyKID] = entry.path
		}
	}

	maps.Copy(conf.Keys, base.Keys)

	conf.KID, _ = chooseSigningKID(entries, signingMargin(base))

	return conf, nil
}

// keySetFromKeyDir builds the key material the issuer should serve: keys on disk sign,
// archived and configured keys verify. The bool reports whether the chosen signer had
// enough headroom
func keySetFromKeyDir(ctx context.Context, base tokens.Config, dir string, configured map[string]crypto.PublicKey, archive keyArchive) (tokens.KeySet, bool, error) {
	entries, err := scanKeyDir(dir)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Str("dir", dir).Msg("failed scanning key directory")

		return tokens.KeySet{}, false, err
	}

	signing := make(map[string]crypto.Signer, len(entries))
	live := make(map[string]crypto.PublicKey, len(entries))

	for _, entry := range entries {
		signing[entry.kid] = entry.signer
		live[entry.kid] = entry.signer.Public()

		// archiving under the legacy kid keeps outstanding tokens verifiable across
		// replicas and after the legacy key file is eventually removed
		if entry.legacyKID != "" {
			live[entry.legacyKID] = entry.signer.Public()
		}
	}

	// record before loading so a key that just appeared comes back in the same pass
	if err := archive.Record(ctx, live); err != nil {
		return tokens.KeySet{}, false, err
	}

	// an archive failure aborts the reload rather than applying a key set without the
	// retired keys, which would stop verifying tokens that are still in flight
	verification, err := archive.Load(ctx)
	if err != nil {
		return tokens.KeySet{}, false, err
	}

	if verification == nil {
		verification = make(map[string]crypto.PublicKey)
	}

	// configured keys keep their historical kid and never sign
	maps.Copy(verification, configured)

	// legacy kids verify directly as well, so they do not depend on the archive round trip
	for _, entry := range entries {
		if entry.legacyKID != "" {
			verification[entry.legacyKID] = entry.signer.Public()
		}
	}

	kid, healthy := chooseSigningKID(entries, signingMargin(base))

	return tokens.KeySet{Signing: signing, Verification: verification, KID: kid}, healthy, nil
}

// configuredVerificationKeys resolves operator configured keys to their public halves once
// at startup, so reloads never depend on the configured paths staying readable and a
// later change to a path's material cannot remap its historical kid
func configuredVerificationKeys(keys map[string]string) (map[string]crypto.PublicKey, error) {
	configured := make(map[string]crypto.PublicKey, len(keys))

	for kid, path := range keys {
		signer, err := tokens.NewFileSigner(path)
		if err != nil {
			return nil, fmt.Errorf("failed loading configured key %q: %w", kid, err)
		}

		configured[kid] = signer.Public()
	}

	return configured, nil
}

// scanKeyDir loads the PEM keys in dir, deriving each kid from the key's thumbprint
func scanKeyDir(dir string) ([]keyDirEntry, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	entries := make([]keyDirEntry, 0, len(files))

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), pemExtension) {
			continue
		}

		name := strings.TrimSuffix(f.Name(), pemExtension)
		path := filepath.Join(dir, f.Name())

		signer, err := tokens.NewFileSigner(path)
		if err != nil {
			return nil, fmt.Errorf("failed loading signing key %q: %w", name, err)
		}

		kid, err := thumbprintKID(signer.Public())
		if err != nil {
			return nil, fmt.Errorf("failed deriving kid for signing key %q: %w", name, err)
		}

		entry := keyDirEntry{kid: kid, path: path, signer: signer}

		if _, err := ulid.Parse(name); err == nil {
			entry.legacyKID = name
		}

		// a key without a cert can still sign, it just has no validity window to
		// schedule a handoff against
		pair, err := tls.LoadX509KeyPair(filepath.Join(dir, name+certExtension), path)

		switch {
		case errors.Is(err, os.ErrNotExist):
		case err != nil:
			return nil, fmt.Errorf("failed loading signing key pair %q: %w", name, err)
		default:
			entry.notBefore = pair.Leaf.NotBefore
			entry.notAfter = pair.Leaf.NotAfter
			entry.hasCert = true
		}

		entries = append(entries, entry)
	}

	if len(entries) == 0 {
		return nil, ErrNoSigningKeys
	}

	return entries, nil
}

// chooseSigningKID returns the kid to sign with and whether it had enough headroom.
// It always returns one of the supplied entries: handing selection back to the token
// manager would let its ULID-preferring fallback promote a verify-only legacy key
func chooseSigningKID(entries []keyDirEntry, margin time.Duration) (string, bool) {
	now := time.Now()

	certBacked := lo.Filter(entries, func(entry keyDirEntry, _ int) bool {
		return entry.hasCert
	})

	if len(certBacked) == 0 {
		return lo.MaxBy(entries, func(a, b keyDirEntry) bool {
			return a.kid > b.kid
		}).kid, true
	}

	eligible := lo.Filter(certBacked, func(entry keyDirEntry, _ int) bool {
		return now.After(entry.notBefore.Add(minCertAge)) && entry.headroom(now) > margin
	})

	mostHeadroom := func(a, b keyDirEntry) bool {
		return a.headroom(now) > b.headroom(now)
	}

	if len(eligible) == 0 {
		return lo.MaxBy(certBacked, mostHeadroom).kid, false
	}

	return lo.MaxBy(eligible, mostHeadroom).kid, true
}

// thumbprintKID derives a stable kid from the RFC 7638 JWK thumbprint of the public key
func thumbprintKID(pub crypto.PublicKey) (string, error) {
	key, err := jwk.Import(pub)
	if err != nil {
		return "", err
	}

	thumbprint, err := key.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(thumbprint), nil
}
