package serveropts

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"maps"

	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/iam/tokens"
)

// WithKeyDir loads all PEM files in dir into the token key map. The filename (without extension)
// is used as the kid. If cert slots are configured, it also loads tls.key/tls.crt pairs and
// uses a public JWK thumbprint as the kid for those keys.
func WithKeyDir(dir string) ServerOption {
	return newApplyFunc(func(s *ServerOptions) {
		if err := applyKeyDir(s, dir); err != nil {
			log.Panic().Err(err).Msg("unable to read key directory")
		}
	})
}

// applyKeyDir reloads signing keys from dir and updates the active token key configuration
func applyKeyDir(s *ServerOptions, dir string) error {
	keys, slots, err := loadKeyDir(s, dir)
	if err != nil {
		return err
	}

	// If no kid found, check config for kid
	if len(keys) == 0 {
		configKid := s.Config.Settings.Auth.Token.KID
		if configKid == "" {
			return fmt.Errorf("no key id (kid) found in key directory and auth.token.kid is not set") //nolint:err113
		}
	}

	if s.configuredTokenKeys == nil {
		s.configuredTokenKeys = maps.Clone(s.Config.Settings.Auth.Token.Keys)
	}

	// Merge only keys that came from static config. Previously discovered keys are rebuilt
	// from the mounted files on every reload so a stale kid cannot survive key rotation.
	maps.Copy(keys, s.configuredTokenKeys)

	s.Config.Settings.Auth.Token.Keys = keys
	if s.Config.Settings.Keywatcher.SignWithCertSlots {
		selectSigningKey(s, slots)
	}

	return nil
}

type certKeySlot struct {
	name        string
	kid         string
	keyPath     string
	certPath    string
	notAfter    time.Time
	renewalTime time.Time
	unsafeAt    time.Time
	safe        bool
}

// loadKeyDir loads legacy root PEM keys and configured cert slots from dir
func loadKeyDir(s *ServerOptions, dir string) (map[string]string, []certKeySlot, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, err
	}

	keys := make(map[string]string)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".pem") {
			continue
		}

		kid := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		keys[kid] = filepath.Join(dir, e.Name())
	}

	slots, err := loadCertSlots(s, dir, keys)
	if err != nil {
		return nil, nil, err
	}

	return keys, slots, nil
}

// loadCertSlots loads cert-manager projected signing key slots and adds them to keys
func loadCertSlots(s *ServerOptions, dir string, keys map[string]string) ([]certKeySlot, error) {
	configuredSlots := s.Config.Settings.Keywatcher.CertSlots
	if len(configuredSlots) == 0 {
		return nil, nil
	}

	slots := make([]certKeySlot, 0, len(configuredSlots))
	maxSigningTTL := maxConfiguredSigningTTL(s)
	safetyWindow := s.Config.Settings.Keywatcher.SafetyWindow

	now := time.Now()

	for _, slotConfig := range configuredSlots {
		if slotConfig.Name == "" {
			return nil, fmt.Errorf("cert slot is missing a name")
		}

		name := slotConfig.Name
		slotDir := filepath.Join(dir, name)
		keyPath := filepath.Join(slotDir, "tls.key")
		certPath := filepath.Join(slotDir, "tls.crt")

		kid, err := keyThumbprintID(keyPath)
		if err != nil {
			return nil, fmt.Errorf("could not derive kid for cert slot %q: %w", name, err)
		}

		slot := certKeySlot{
			name:     name,
			kid:      kid,
			keyPath:  keyPath,
			certPath: certPath,
		}

		if certPath != "" && slotConfig.RenewBefore > 0 {
			notAfter, err := certificateNotAfter(certPath)
			if err != nil {
				return nil, fmt.Errorf("could not read certificate for cert slot %q: %w", name, err)
			}

			slot.notAfter = notAfter
			slot.renewalTime = notAfter.Add(-slotConfig.RenewBefore)
			slot.unsafeAt = slot.renewalTime.Add(-maxSigningTTL).Add(-safetyWindow)
			slot.safe = now.Before(slot.unsafeAt)
		}

		keys[kid] = keyPath
		slots = append(slots, slot)
	}

	return slots, nil
}

// keyThumbprintID derives a stable JWK thumbprint kid from the slot private key public component
func keyThumbprintID(keyPath string) (string, error) {
	signer, err := tokens.NewFileSigner(keyPath)
	if err != nil {
		return "", err
	}

	key, err := jwk.Import(signer.Public())
	if err != nil {
		return "", err
	}

	thumbprint, err := key.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(thumbprint), nil
}

// certificateNotAfter returns the NotAfter timestamp from the first certificate PEM block
func certificateNotAfter(certPath string) (time.Time, error) {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return time.Time{}, err
	}

	for {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return time.Time{}, err
		}

		return cert.NotAfter, nil
	}

	return time.Time{}, fmt.Errorf("no certificate PEM block found in %s", certPath) //nolint:err113
}

// maxConfiguredSigningTTL returns the configured maximum lifetime for tokens signed by a key
func maxConfiguredSigningTTL(s *ServerOptions) time.Duration {
	if ttl := s.Config.Settings.Keywatcher.MaxSigningTTL; ttl > 0 {
		return ttl
	}

	tokenConfig := s.Config.Settings.Auth.Token
	maxTTL := tokenConfig.AccessDuration

	for _, ttl := range []time.Duration{
		tokenConfig.RefreshDuration,
		tokenConfig.AssessmentAccessDuration,
		tokenConfig.TrustCenterNDARequestAccessDuration,
	} {
		if ttl > maxTTL {
			maxTTL = ttl
		}
	}

	if maxTTL <= 0 {
		return 168 * time.Hour //nolint:mnd
	}

	return maxTTL
}

// selectSigningKey chooses a safe cert slot for signing and updates the configured kid
func selectSigningKey(s *ServerOptions, slots []certKeySlot) {
	if len(slots) == 0 {
		return
	}

	currentKID := s.Config.Settings.Auth.Token.KID
	for _, slot := range slots {
		if slot.kid == currentKID && slot.safe {
			log.Info().Str("kid", slot.kid).Str("slot", slot.name).Time("unsafe_at", slot.unsafeAt).Msg("keeping current jwt signing key")
			return
		}
	}

	safeSlots := make([]certKeySlot, 0, len(slots))
	for _, slot := range slots {
		if slot.safe {
			safeSlots = append(safeSlots, slot)
		}
	}

	if len(safeSlots) == 0 {
		log.Error().Str("current_kid", currentKID).Msg("no safe cert slot available for jwt signing; keeping configured signing key")
		return
	}

	sort.Slice(safeSlots, func(i, j int) bool {
		if safeSlots[i].unsafeAt.Equal(safeSlots[j].unsafeAt) {
			return safeSlots[i].kid < safeSlots[j].kid
		}

		return safeSlots[i].unsafeAt.After(safeSlots[j].unsafeAt)
	})

	selected := safeSlots[0]
	s.Config.Settings.Auth.Token.KID = selected.kid

	log.Info().Str("kid", selected.kid).Str("slot", selected.name).Time("unsafe_at", selected.unsafeAt).Msg("selected jwt signing key")
}
